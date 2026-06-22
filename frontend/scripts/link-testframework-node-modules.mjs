/**
 * Symlink frontend/node_modules for testframework E2E + Vitest paths outside frontend/.
 */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const frontendNodeModules = path.resolve(__dirname, "../node_modules");
const targets = [
	path.resolve(__dirname, "../../testframework/stages/e2e/node_modules"),
	path.resolve(__dirname, "../../testframework/stages/unit/frontend-test/node_modules"),
];

if (!fs.existsSync(frontendNodeModules)) {
	console.error("link-testframework-node-modules: run npm ci in frontend/ first");
	process.exit(1);
}

for (const target of targets) {
	if (fs.existsSync(target)) {
		const stat = fs.lstatSync(target);
		if (stat.isSymbolicLink() || stat.isDirectory()) {
			continue;
		}
	}
	fs.mkdirSync(path.dirname(target), { recursive: true });
	const type = process.platform === "win32" ? "junction" : "dir";
	fs.symlinkSync(frontendNodeModules, target, type);
	console.log("linked", target, "->", frontendNodeModules);
}
