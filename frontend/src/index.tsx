/* @refresh reload */
import { render } from "solid-js/web";
import { ColorModeProvider, ColorModeScript, createLocalStorageManager } from "@kobalte/core";
import "./app.css";
import App from "./App.tsx";

const root = document.getElementById("root");
const storageManager = createLocalStorageManager("ariadm-theme");

render(() => (
  <>
    <ColorModeScript storageType={storageManager.type} />
    <ColorModeProvider storageManager={storageManager}>
      <App />
    </ColorModeProvider>
  </>
), root!);
