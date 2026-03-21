/** Plate JSON renderer for message content. */

/** Render one Plate JSON node. */
export function RenderNode({ node }: { node: Record<string, unknown> }) {
  // Text node.
  if ("text" in node && typeof node.text === "string") {
    let el: React.ReactNode = node.text;
    if (node.bold) el = <strong>{el}</strong>;
    if (node.italic) el = <em>{el}</em>;
    if (node.code)
      el = (
        <code className="rounded bg-muted px-1.5 py-0.5 text-xs font-mono">
          {el}
        </code>
      );
    if (node.strikethrough) el = <s>{el}</s>;
    return <>{el}</>;
  }

  const children =
    (node.children as Record<string, unknown>[] | undefined) ?? [];
  const renderedChildren = children.map((child, i) => (
    <RenderNode key={i} node={child} />
  ));

  switch (node.type) {
    case "p":
      return <p className="mb-1 leading-relaxed">{renderedChildren}</p>;
    case "h1":
      return <h1 className="text-lg font-bold">{renderedChildren}</h1>;
    case "h2":
      return <h2 className="text-base font-bold">{renderedChildren}</h2>;
    case "ul":
      return <ul className="list-disc pl-4">{renderedChildren}</ul>;
    case "ol":
      return <ol className="list-decimal pl-4">{renderedChildren}</ol>;
    case "li":
      return <li>{renderedChildren}</li>;
    case "blockquote":
      return (
        <blockquote className="border-l-2 border-primary/30 pl-3 italic text-muted-foreground">
          {renderedChildren}
        </blockquote>
      );
    case "a":
      return (
        <a
          href={node.url as string}
          className="text-primary underline decoration-primary/30 hover:decoration-primary"
          target="_blank"
          rel="noopener noreferrer"
        >
          {renderedChildren}
        </a>
      );
    case "code_block":
      return (
        <pre className="rounded-lg bg-muted p-3 text-xs font-mono overflow-x-auto">
          <code>{renderedChildren}</code>
        </pre>
      );
    default:
      return <>{renderedChildren}</>;
  }
}

/** Render message content from Plate JSON (simplified). */
export function RenderContent({
  contentJson,
}: {
  contentJson?: Record<string, unknown>;
}) {
  if (!contentJson) return null;

  const children = (contentJson as { children?: unknown[] }).children;
  if (!Array.isArray(children)) return null;

  return (
    <div className="prose prose-sm dark:prose-invert max-w-none">
      {children.map((node, i) => (
        <RenderNode key={i} node={node as Record<string, unknown>} />
      ))}
    </div>
  );
}
