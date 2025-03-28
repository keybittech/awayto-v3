export function generateLightBgColor(): string {
  // Function to convert a single component of a color from hex to decimal
  function hexToDec(hex: string): number {
    return parseInt(hex, 16);
  }

  // Function to calculate the relative luminance of a color
  function calculateLuminance(r: number, g: number, b: number): number {
    let a = [r, g, b].map(v => {
      v /= 255;
      return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4);
    });
    return a[0] * 0.2126 + a[1] * 0.7152 + a[2] * 0.0722;
  }

  // Function to generate a random light color
  function generateColor(): string {
    let color = Math.floor(Math.random() * Math.pow(256, 3)).toString(16);
    while (color.length < 6) {
      color = '0' + color;
    }
    return color;
  }

  // Generate a random color
  let color = generateColor();

  // Calculate the relative luminance of the color
  let luminance = calculateLuminance(
    hexToDec(color.slice(0, 2)),
    hexToDec(color.slice(2, 4)),
    hexToDec(color.slice(4, 6))
  );

  // If the color is too dark, generate a new color
  while (luminance < 0.5) {
    color = generateColor();
    luminance = calculateLuminance(
      hexToDec(color.slice(0, 2)),
      hexToDec(color.slice(2, 4)),
      hexToDec(color.slice(4, 6))
    );
  }

  // Return the color
  return '#' + color;
}
