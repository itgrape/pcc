import { useEffect, useMemo, useState } from "react"
import { autocompletion, type CompletionContext } from "@codemirror/autocomplete"
import { json } from "@codemirror/lang-json"
import { yaml } from "@codemirror/lang-yaml"
import { StreamLanguage } from "@codemirror/language"
import { properties } from "@codemirror/legacy-modes/mode/properties"
import { oneDark } from "@codemirror/theme-one-dark"
import CodeMirror from "@uiw/react-codemirror"
import { Braces, CheckCircle2, Sparkles, TriangleAlert } from "lucide-react"
import { parse as parseYaml, stringify as stringifyYaml } from "yaml"

import { Button } from "@/components/ui/button"
import { useTheme } from "@/features/theme/theme-context"
import { cn } from "@/lib/utils"

interface ConfigEditorProps {
  value: string
  onChange: (value: string) => void
  format: number
  height?: number
  onValidityChange?: (valid: boolean) => void
}

const languages: Record<number, { id: string; label: string }> = {
  1: { id: "json", label: "JSON" },
  2: { id: "yaml", label: "YAML" },
  3: { id: "plaintext", label: "Text" },
  4: { id: "ini", label: "Properties" },
}

function formatContent(value: string, format: number) {
  if (!value.trim()) {
    return value
  }
  if (format === 1) {
    return `${JSON.stringify(JSON.parse(value), null, 2)}\n`
  }
  if (format === 2) {
    return stringifyYaml(parseYaml(value), { indent: 2, lineWidth: 100 })
  }
  return value
    .split("\n")
    .map((line) => line.trimEnd())
    .join("\n")
}

function validateContent(value: string, format: number) {
  if (!value.trim() || (format !== 1 && format !== 2)) {
    return null
  }
  try {
    if (format === 1) {
      JSON.parse(value)
    } else {
      parseYaml(value)
    }
    return null
  } catch (error) {
    return error instanceof Error ? error.message : "配置内容格式不正确"
  }
}

function configCompletions(context: CompletionContext) {
  const word = context.matchBefore(/[\w.-]*/)
  if (!word || (word.from === word.to && !context.explicit)) {
    return null
  }
  return {
    from: word.from,
    options: [
      { label: "true", type: "keyword" },
      { label: "false", type: "keyword" },
      { label: "null", type: "keyword" },
      { label: "localhost", type: "text" },
    ],
  }
}

function languageExtension(format: number) {
  switch (format) {
    case 1:
      return json()
    case 2:
      return yaml()
    case 4:
      return StreamLanguage.define(properties)
    default:
      return []
  }
}

export function ConfigEditor({
  value,
  onChange,
  format,
  height = 360,
  onValidityChange,
}: ConfigEditorProps) {
  const { theme } = useTheme()
  const [formatError, setFormatError] = useState<string | null>(null)
  const language = languages[format] ?? languages[3]
  const extensions = useMemo(
    () => [languageExtension(format), autocompletion({ override: [configCompletions] })],
    [format],
  )
  const validationError = useMemo(() => validateContent(value, format), [value, format])
  const error = formatError ?? validationError

  useEffect(() => {
    onValidityChange?.(!error)
  }, [error, onValidityChange])

  const handleFormat = () => {
    try {
      onChange(formatContent(value, format))
      setFormatError(null)
    } catch (formattingError) {
      setFormatError(
        formattingError instanceof Error ? formattingError.message : "无法格式化配置内容",
      )
    }
  }

  return (
    <div className="overflow-hidden rounded-2xl border bg-[#fbfbfd] shadow-inner dark:bg-[#0d0f17]">
      <div className="flex items-center justify-between border-b bg-muted/45 px-3 py-2">
        <div className="flex items-center gap-2 text-xs font-medium">
          <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <Braces className="h-3.5 w-3.5" />
          </span>
          <span>{language.label}</span>
          <span className="text-muted-foreground">支持语法高亮、自动补全与校验</span>
        </div>
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "flex items-center gap-1.5 text-xs",
              error ? "text-destructive" : "text-emerald-600 dark:text-emerald-400",
            )}
          >
            {error ? (
              <TriangleAlert className="h-3.5 w-3.5" />
            ) : (
              <CheckCircle2 className="h-3.5 w-3.5" />
            )}
            {error ? "格式有误" : "格式有效"}
          </div>
          <Button type="button" variant="outline" size="sm" onClick={handleFormat}>
            <Sparkles className="h-3.5 w-3.5" />
            一键格式化
          </Button>
        </div>
      </div>
      <CodeMirror
        height={`${height}px`}
        value={value}
        theme={theme === "dark" ? oneDark : "light"}
        extensions={extensions}
        onChange={(nextValue) => {
          setFormatError(null)
          onChange(nextValue)
        }}
        basicSetup={{
          autocompletion: true,
          bracketMatching: true,
          closeBrackets: true,
          foldGutter: true,
          highlightActiveLine: true,
          highlightActiveLineGutter: true,
          indentOnInput: true,
          lineNumbers: true,
        }}
        className="text-[13px] [&_.cm-content]:font-mono [&_.cm-content]:leading-[21px] [&_.cm-editor]:outline-none [&_.cm-gutters]:border-r [&_.cm-gutters]:border-border/60"
      />
      {error && (
        <div className="border-t border-destructive/20 bg-destructive/5 px-4 py-2 text-xs text-destructive">
          {error}
        </div>
      )}
    </div>
  )
}
