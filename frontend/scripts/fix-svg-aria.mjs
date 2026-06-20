import fs from "node:fs";
import path from "node:path";

function walk(dir, acc = []) {
	for (const entry of fs.readdirSync(dir)) {
		const full = path.join(dir, entry);
		const stat = fs.statSync(full);
		if (stat.isDirectory() && entry !== "node_modules") {
			walk(full, acc);
		} else if (/\.(tsx|jsx)$/.test(entry)) {
			acc.push(full);
		}
	}
	return acc;
}

let fixed = 0;
for (const file of walk("src")) {
	let content = fs.readFileSync(file, "utf8");
	let changed = false;
	content = content.replace(/<svg\b([^>]*)>/g, (match, attrs) => {
		if (/aria-hidden|aria-label|role\s*=/.test(attrs)) {
			return match;
		}
		fixed += 1;
		changed = true;
		const trimmed = attrs.trimStart();
		return trimmed
			? `<svg aria-hidden="true" ${trimmed}>`
			: '<svg aria-hidden="true">';
	});
	if (changed) {
		fs.writeFileSync(file, content);
	}
}

console.log(`fixed ${fixed} svg tags`);
