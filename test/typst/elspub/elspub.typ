#import "@preview/articulate-coderscompass:0.1.7": *

#set text(
  font: "Lato",
  style: "normal",
  size: 10pt,
)

#let ctx = json("test_data.json")

#show: articulate-coderscompass.with(
  title: ctx.title,
  subtitle: ctx.subtitle,
  authors: (
    ctx.authors.map(a => (name: a.name, email: a.email, affiliation: a.affiliation))
  ),
  abstract: ctx.abstract,
  keywords: (),
  version: "1.0.0",
  reading-time: "6 minutes",
  date: datetime.today(),
)

#render-markdown(read("md_content/content.md"))

// Or write Typst directly

= Article Title

Content here.
