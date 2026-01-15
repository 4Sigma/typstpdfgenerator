#import "@preview/plotst:0.2.0": *

#set page(paper: "a4", margin: 2cm)
#set text(font: "Liberation Sans")

#align(center)[
  #text(size: 20pt, weight: "bold")[
    Data Visualization Test
  ]
]

#v(1em)

= Introduction

This document demonstrates data visualization capabilities using the plotst package.

= Line Plot Example

#plot(
  data: (
    ([0, 1, 2, 3, 4, 5], [0, 1, 4, 9, 16, 25]),
  ),
  size: (12, 8),
  x-label: "X Axis",
  y-label: "Y Axis",
  caption: [Quadratic function: $y = x^2$],
)

= Bar Chart Example

#bar-chart(
  data: (
    ([Jan, Feb, Mar, Apr, May], [10, 15, 13, 17, 20]),
  ),
  size: (12, 8),
  x-label: "Month",
  y-label: "Sales (k€)",
  caption: [Monthly sales data],
)

= Scatter Plot Example

#scatter-plot(
  data: (
    ([1, 2, 3, 4, 5, 6, 7, 8], [2.1, 3.9, 6.2, 8.1, 9.8, 12.3, 14.1, 16.2]),
  ),
  size: (12, 8),
  x-label: "Input",
  y-label: "Output",
  caption: [Experimental data points],
)

= Conclusion

Charts and graphs are rendered correctly by the remote FaaS service.
