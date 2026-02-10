const _colors = [
  "53, 158, 255",
  "96, 122, 251",
  "57, 224, 121",
  "208, 187, 149",
  "76, 22, 201",
  "234, 40, 49",
  "250, 198, 56",
]

export function getColor(index: number): string {
  return _colors[index % _colors.length]!
}
