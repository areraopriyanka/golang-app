// Styling

#let filledCellColor = luma(90%)

#set outline.entry(
  fill: none,
)

#set par(spacing: 30pt, leading: 8pt)

#set text(font: "Inter", size: 8.2pt, weight: "regular")

#set strong(delta: 300)

#set list(
  spacing: 12pt,
  indent: 8pt,
  marker: ([•], [◦], [▪]),
)

#set table(inset: 8pt, stroke: (paint: luma(50%)))

#show table.cell: c => {
  set par(spacing: 15pt, leading: 8pt)
  set list(spacing: 8pt)
  if c.fill != none {
    set text(size: 8.2pt, weight: "semibold")
  }
  c
}

#show list: l => {
  set par(spacing: 12pt, leading: 8pt)
  l
}

#show link: l => {
  underline(text(
    l,
    weight: "bold",
    blue,
  ))
}

#show heading: h => {
  block(above: 40pt, below: 15pt, text(
    h.body,
    size: 8.2pt,
    weight: "semibold",
  ))
}

// Elements

#let nodeText(node) = {
  if node.at("bold", default: false) {
    strong(text(node.text))
  } else {
    text(node.text)
  }
}

#let nodeLink(node) = {
  link(node.at("link", default: node.text), node.text)
}

// Displays: Text, Link
#let contentInlineText(node) = {
  if node.type == "text" {
    nodeText(node)
  } else if node.type == "link" {
    nodeLink(node)
  } else {
    panic("Unexpected element type in JSON", node.type)
  }
}

#let nodeParagraph(node) = {
  node.content.map(contentInlineText).join()
}

// Displays: Paragraph, Text, Link
#let contentText(node) = {
  if node.type == "paragraph" {
    nodeParagraph(node)
  } else {
    contentInlineText(node)
  }
}

// Displays: List, Paragraph, Text, Link
#let contentList(node) = {
  if type(node) == array {
    // This list entry has multiple parts
    node.map(contentList).join(parbreak())
  } else if node.type == "list" {
    // Recursively break down lists
    list(..node.content.map(contentList))
  } else {
    contentText(node)
  }
}

#let nodeTableCell(cell, colspan) = {
  let headerText = cell.at("header", default: none)
  if headerText != none {
    table.cell(
      text(
        headerText,
        size: 8.2pt,
        weight: "semibold",
      ),
      fill: filledCellColor,
      colspan: colspan,
    )
  } else {
    table.cell(
      cell.content.map(contentList).join(parbreak()),
      colspan: colspan,
    )
  }
}

#let nodeTableRow(row, columnCount) = {
  for (index, cell) in row.enumerate() {
    let colspan = if index == row.len() - 1 {
      calc.max(1, columnCount - index)
    } else { 1 }

    (nodeTableCell(cell, colspan),)
  }
}

#let nodeTableColumnFraction(column) = {
  column * 1fr
}

#let nodeTable(node) = {
  table(
    columns: node.columns.map(nodeTableColumnFraction),
    ..for row in node.rows {
      (..nodeTableRow(row, node.columns.len()),)
    }
  )
}

#let nodeHeader(node) = [
  = #node.text
]

// Displays: Table, Header, List, Paragraph, Text, Link
#let contentTable(node) = {
  if node.type == "table" {
    nodeTable(node)
  } else if node.type == "header" {
    nodeHeader(node)
  } else {
    contentList(node)
  }
}

#let nodeCollapsible(node) = [
  = #node.header
  #node.content.map(contentTable).join(parbreak())
]

#let nodeSpacer(node) = []

// Displays: Collapsible, Spacer, Table, Header, List, Paragraph, Text, Link
#let contentCollapsible(node) = {
  if node.type == "collapsible" {
    nodeCollapsible(node)
  } else if node.type == "spacer" {
    nodeSpacer(node)
  } else {
    contentTable(node)
  }
}

// Render the agreement from JSON

#let formatDate((year, month, day)) = {
  let date = datetime(year: year, month: month, day: day)
  date.display("[month repr:long] [day], [year]")
}

#let agreement(data) = {
  block(
    below: 20pt,
    text(
      data.title,
      size: 14.3pt,
      weight: "regular",
    ),
  )

  [#data.at("updatedText", default: "Last Update"): #formatDate(data.updatedAt)]

  if data.at("outline", default: false) {
    outline(
      title: "Table of Contents",
    )
  }

  data.content.map(contentCollapsible).join(parbreak())
}

// This value should be set by compiling this file with the input parameter:
//   --input file="agreement.json"
//
// A default value can be specified here to allow for LSP type checking.
#let data = json(sys.inputs.at("file"))

#agreement(data)
