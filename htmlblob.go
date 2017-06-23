package main

import (
	"bytes"
)

type CommentBlob struct {
	bytes.Buffer
}

func (b *CommentBlob) Open() {
	b.WriteString("<ul id=comment_blob>")
}

func (b *CommentBlob) Append(name string, comment string) {
	b.WriteString("<li>")
	b.WriteString("<div>")
	b.WriteString(name)
	b.WriteString("</div>")
	b.WriteString("<div>")
	b.WriteString(comment)
	b.WriteString("</div>")
	b.WriteString("</li>")
}

func (b *CommentBlob) Close() {
	b.WriteString("</ul>")
}
