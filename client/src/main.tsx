import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";

// vConsole（モバイルデバッグ用）
if (import.meta.env.VITE_VCONSOLE_ENABLED === "true") {
  import("vconsole").then((module) => {
    new module.default();
  });
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
