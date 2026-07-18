const fs = require("fs");
const isHex = c => (c>="0"&&c<="9")||(c>="a"&&c<="f")||(c>="A"&&c<="F");
let raw = fs.readFileSync("C:/Users/Valerii/Desktop/hub/.planning/biome.json", "utf8");
// Fix invalid backslash escapes everywhere — replace any backslash inside a
// double-quoted string that isn't part of a valid JSON escape (" \ / b f n r t u).
// Approach: walk the string char-by-char with a tiny state machine.
let out = "";
let inStr = false;
for (let i = 0; i < raw.length; i++) {
  const ch = raw[i];
  if (!inStr) {
    if (ch === '"') inStr = true;
    out += ch;
  } else {
    if (ch === '"') { inStr = false; out += ch; continue; }
    if (ch === "\\") {
      const next = raw[i + 1];
      if (next === '"' || next === "\\" || next === "/") {
        out += ch + next;
        i++;
      } else if (next === "u" && isHex(raw[i+2]) && isHex(raw[i+3]) && isHex(raw[i+4]) && isHex(raw[i+5])) {
        out += ch + next;
        i++;
      } else {
        // invalid escape, double it
        out += "\\\\";
      }
    } else {
      out += ch;
    }
  }
}
raw = out;
const d = JSON.parse(raw);
const cat = process.argv[2] || "lint/suspicious/noArrayIndexKey";
const arr = d.diagnostics.filter(x => x.category === cat);
console.log("count:", arr.length);
for (const x of arr) {
  const p = x.location && x.location.path;
  const pathStr = typeof p === "string" ? p : (p && p.file) || JSON.stringify(p);
  console.log(pathStr + ":" + x.location.start.line + ":" + x.location.start.column + (x.message ? "  -- " + x.message.slice(0,80) : ""));
}
