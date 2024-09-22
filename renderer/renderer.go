package renderer

import (
	"fmt"
	"io"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func renderImage(
	w io.Writer,
	node *ast.Image,
	entering bool,
	dirpath *string,
	fn func(string) string,
) {
	if entering {
		// TODO: point to path correctly, not using local ip address
		htmlPath := fmt.Sprintf(
			"http://192.168.0.166:4041/images/%s/%s",
			*dirpath, fn(string(node.Destination)),
		)
		w.Write(
			[]byte(fmt.Sprintf(
				"<img src=%s alt=%s>",
				htmlPath,
				string(node.Title),
			)),
		)
	} else {
		w.Write([]byte("</img>"))
	}
}

func makeRenderHook(
	dirpath *string,
	fn func(string) string,
) html.RenderNodeFunc {
	return func(
		w io.Writer,
		node ast.Node,
		entering bool,
	) (ast.WalkStatus, bool) {
		if image, ok := node.(*ast.Image); ok {
			renderImage(w, image, entering, dirpath, fn)
			return ast.GoToNext, true
		}

		return ast.GoToNext, false
	}
}

func MdToHTML(md []byte, dirpath *string, fn func(string) string) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions |
		parser.AutoHeadingIDs |
		parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{
		Flags:          htmlFlags,
		RenderNodeHook: makeRenderHook(dirpath, fn),
	}
	renderer := html.NewRenderer(opts) // TODO: singleton (?)

	return markdown.Render(doc, renderer)
}
