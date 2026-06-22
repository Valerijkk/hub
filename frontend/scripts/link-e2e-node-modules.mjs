/**
 * Symlink frontend/node_modules → testframework/e2e/node_modules so E2E specs
 * (outside frontend/) resolve @playwright/test and @axe-core/playwright.
 */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const frontendNodeModules = path.resolve(__dirname, "../node_modules");
const e2eNodeModules = path.resolve(__dirname, "../../testframework/e2e/node_modules");

if (!fs.existsSync(frontendNodeModules)) {
	console.error("link-e2e-node-modules: run npm ci in frontend/ first");
	process.exit(1);
}

if (fs.existsSync(e2eNodeModules)) {
	const stat = fs.lstatSync(e2eNodeModules);
	if (stat.isSymbolicLink() || stat.isDirectory()) {
		process.exit(0);
	}
}

fs.mkdirSync(path.dirname(e2eNodeModules), { recursive: true });
const type = process.platform === "win32" ? "junction" : "dir";
fs.symlinkSync(frontendNodeModules, e2eNodeModules, type);
console.log("linked", e2eNodeModules, "->", frontendNodeModules);
