package renderer

import (
    "bytes"
	"fmt"
	"io"
    "strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var gallery = []byte(":gallery\n")

type Gallery struct {
	ast.Leaf
	ImageURLS []string
}

func genHtmlPath(dirpath string, name string, sanitize func(string) string) string {
    return fmt.Sprintf("/images/%s/%s", dirpath, sanitize(name))
}

func renderImage(
	w io.Writer,
	node *ast.Image,
	entering bool,
	dirpath *string,
	fn func(string) string,
) {
    // leaf node, no else/leaving
	if entering {
        htmlPath := genHtmlPath(*dirpath, string(node.Destination), fn)
        io.WriteString(w, fmt.Sprintf("<img src=%s alt=%s>", htmlPath, string(node.Title)))
	} else {
        io.WriteString(w, "</img>")
    }
}

func renderGallery(
	w io.Writer,
    g *Gallery,
	entering bool,
	dirpath *string,
	fn func(string) string,
) {
    if entering {
        io.WriteString(w, "\n<gallery>")

        for i, name := range g.ImageURLS {
            htmlPath := genHtmlPath(*dirpath, name, fn)
            io.WriteString(
                w,
                fmt.Sprintf(
                    "\n<img src=%s alt=%s>\n",
                    htmlPath,
                    fmt.Sprintf("image-%d-%d", i/2, i%2),
                ),
            )
            io.WriteString(w, "</img>")
        }

        io.WriteString(w, "<gallery>\n")
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

        if gallery, ok := node.(*Gallery); ok {
            renderGallery(w, gallery, entering, dirpath, fn)
            return ast.GoToNext, true
        }

		return ast.GoToNext, false
	}
}

func parseGallery(data []byte) (ast.Node, []byte, int) {
	if !bytes.HasPrefix(data, gallery) {
		return nil, nil, 0
	}

	i := len(gallery)
	// find empty line
	// TODO: should also consider end of document
	end := bytes.Index(data[i:], []byte("\n\n"))
	if end < 0 {
		return nil, data, 0
	}

	end = end + i
	lines := string(data[i:end])
	parts := strings.Split(lines, "\n")
	res := &Gallery{
		ImageURLS: parts,
	}

	return res, nil, end
}

func parserHook(data []byte) (ast.Node, []byte, int) {
	if node, d, n := parseGallery(data); node != nil {
		return node, d, n
	}

	return nil, nil, 0
}

func MdToHTML(md []byte, dirpath *string, fn func(string) string) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions |
		parser.AutoHeadingIDs |
		parser.NoEmptyLineBeforeBlock ^
        parser.DefinitionLists
	p := parser.NewWithExtensions(extensions)
    p.Opts.ParserHook = parserHook
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{
		Flags:          htmlFlags,
		RenderNodeHook: makeRenderHook(dirpath, fn),
	}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
