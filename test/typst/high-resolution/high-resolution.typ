#set page(paper: "a4", margin: 2cm)
#set text(font: "Liberation Sans")

#align(center)[
  #text(size: 20pt, weight: "bold")[
    High Resolution Test
  ]
]

#v(1em)

This document was generated with `--ppi 300` option for high resolution output.

= Graphics and Images

High resolution settings are particularly important when the PDF contains:

- Rasterized graphics
- Embedded images
- Complex vector graphics
- Charts and diagrams

== Quality Comparison

#table(
  columns: 3,
  [Setting], [PPI], [Use Case],
  [Low], [72], [Screen preview],
  [Medium], [144], [General purpose (default)],
  [High], [300], [Print quality],
  [Very High], [600], [Professional printing],
)

#v(1em)

#align(center)[
  #rect(
    width: 80%,
    height: 3cm,
    fill: gradient.linear(..color.map.rainbow),
    stroke: 1pt,
  )

  _High resolution gradient test_
]
