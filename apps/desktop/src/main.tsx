
import ReactDOM from "react-dom/client";
import App from "./DesktopApp";
import { installDesktopAuthFetchInterceptor } from "./api/auth";
import { applyTheme, resolveInitialTheme } from "./design-system";
import "./design-system/styles.css";

installDesktopAuthFetchInterceptor();
applyTheme(resolveInitialTheme());

const root = document.getElementById("root");

if (root) {
  ReactDOM.createRoot(root).render(<App />);
}
