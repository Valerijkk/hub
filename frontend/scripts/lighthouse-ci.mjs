#!/usr/bin/env node
/**
 * FB-06: Lighthouse CI helper — accessibility/performance smoke.
 * Usage: npm run build && npm run preview & npm run lighthouse:ci
 */
import { spawn } from "node:child_process";
import { readFileSync, writeFileSync } from "node:fs";

const URL = process.env.LH_URL ?? "http://localhost:4173/login";
const THRESHOLDS = {
	accessibility: 0.9,
	performance: 0.75,
	"best-practices": 0.9,
};

const child = spawn(
	"npx",
	[
		"lighthouse",
		URL,
		"--only-categories=accessibility,performance,best-practices",
		"--output=json",
		"--output-path=./lighthouse-report.json",
		"--chrome-flags=--headless --no-sandbox",
		"--quiet",
	],
	{ stdio: "inherit", shell: true },
);

child.on("close", (code) => {
	if (code !== 0) {
		console.error("Lighthouse failed");
		process.exit(code ?? 1);
	}

	const report = JSON.parse(readFileSync("./lighthouse-report.json", "utf8"));
	const summary = {};
	let failed = false;

	for (const [cat, minScore] of Object.entries(THRESHOLDS)) {
		const score = report.categories[cat]?.score ?? 0;
		summary[cat] = score;
		if (score < minScore) {
			failed = true;
			console.error(`${cat}: ${(score * 100).toFixed(0)}% < ${minScore * 100}%`);
		} else {
			console.log(`${cat}: ${(score * 100).toFixed(0)}% OK`);
		}
	}

	writeFileSync("./lighthouse-summary.json", JSON.stringify(summary, null, 2));
	process.exit(failed ? 1 : 0);
});
