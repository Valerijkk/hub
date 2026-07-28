/** Текстовый узел Plate JSON. */
export interface PlateTextNode {
	text: string;
	bold?: boolean;
	italic?: boolean;
	code?: boolean;
}

/** Элементный узел Plate JSON. */
export interface PlateElementNode {
	type: string;
	children: (PlateTextNode | PlateElementNode)[];
	url?: string;
}

/** Узкий Plate-документ для content_json (без double-cast). */
export interface PlateDoc {
	type: "doc";
	children: PlateElementNode[];
}

/** Преобразует PlateDoc в Record для SendMessageRequest без `as unknown as`. */
export function plateDocToContentJson(doc: PlateDoc): Record<string, unknown> {
	return { type: doc.type, children: doc.children };
}
