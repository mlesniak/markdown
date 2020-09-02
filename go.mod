module github.com/mlesniak/markdown

go 1.13

require (
	github.com/labstack/echo/v4 v4.1.15
	github.com/labstack/gommon v0.3.0
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/ziflex/lecho/v2 v2.0.0
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073 // indirect
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
)

// replace github.com/russross/blackfriday/v2 => github.com/mlesniak/blackfriday/v2 v2.0.1
