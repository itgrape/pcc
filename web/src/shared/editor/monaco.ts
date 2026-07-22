import { loader } from "@monaco-editor/react"
import * as monaco from "monaco-editor/esm/vs/editor/editor.api.js"
import "monaco-editor/esm/vs/language/json/monaco.contribution.js"
import "monaco-editor/esm/vs/basic-languages/monaco.contribution.js"
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker.js?worker"
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker.js?worker"

self.MonacoEnvironment = {
  getWorker(_moduleId: string, label: string) {
    if (label === "json") {
      return new jsonWorker()
    }
    return new editorWorker()
  },
}

loader.config({ monaco })
