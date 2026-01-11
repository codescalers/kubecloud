export function toPrecision(value: number, precision: number) {
  return Math.round(value * 10 ** precision) / 10 ** precision
}
