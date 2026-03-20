import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./app.css";

const root = document.getElementById("root");
if (!root) {
  throw new Error("Root element not found");
}

createRoot(root).render(
  <StrictMode>
    <div>
      <h1>Hub</h1>
    </div>
  </StrictMode>,
);
