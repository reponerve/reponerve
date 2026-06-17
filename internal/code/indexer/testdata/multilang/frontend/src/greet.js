export function greet(name) {
  return `hello ${name}`;
}

export class Greeter {
  say(name) {
    return greet(name);
  }
}
