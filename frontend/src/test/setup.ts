import "@testing-library/jest-dom/vitest";

if (typeof Element.prototype.scrollTo !== "function") {
	Object.defineProperty(Element.prototype, "scrollTo", {
		configurable: true,
		writable: true,
		value: () => undefined,
	});
}

if (typeof Element.prototype.scrollIntoView !== "function") {
	Object.defineProperty(Element.prototype, "scrollIntoView", {
		configurable: true,
		writable: true,
		value: () => undefined,
	});
}
