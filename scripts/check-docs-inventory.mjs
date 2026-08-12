#!/usr/bin/env node
/**
 * TASK-080 / TASK-054: assert docs inventory counts match measured tree.
 * Mutation: change claimed backend test count in docs → exit 1.
 */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const tfDir = path.join(root, "backend", "testframework");

function countTestGo(dir) {
	let n = 0;
	for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
		const p = path.join(dir, ent.name);
		if (ent.isDirectory()) n += countTestGo(p);
		else if (ent.isFile() && ent.name.endsWith("_test.go")) n += 1;
	}
	return n;
}

const measured = countTestGo(tfDir);
const docs = [
	"docs/TROUBLESHOOTING.md",
	"docs/test_plan.md",
	"docs/TESTING.md",
	"docs/testframework_strategy.md",
];

let drift = false;
for (const rel of docs) {
	const full = path.join(root, rel);
	const text = fs.readFileSync(full, "utf8");
	const claims = [];
	for (const line of text.split(/\r?\n/)) {
		if (!/testframework|_test\.go|тестовых файлов/i.test(line)) continue;
		if (/export_hooks/i.test(line)) continue;
		for (const m of line.matchAll(/\*\*(\d+)\*\*/g)) {
			const n = Number(m[1]);
			// ignore tiny / unrelated counts on the same inventory tables
			if (n < 100) continue;
			claims.push(n);
		}
	}
	if (claims.length === 0) {
		console.error(`FAIL ${rel}: no inventory claim matched`);
		drift = true;
		continue;
	}
	for (const c of claims) {
		if (c !== measured) {
			console.error(
				`FAIL ${rel}: claimed ${c} _test.go, measured ${measured}`,
			);
			drift = true;
		} else {
			console.log(`OK ${rel}: ${c} == measured ${measured}`);
		}
	}
}

if (drift) {
	console.error(`docs inventory drift (measured=${measured})`);
	process.exit(1);
}
console.log(`docs inventory OK (measured=${measured})`);
