<template>
  <div ref="host" class="w-full h-full min-h-0"></div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import '../core/monacoEnv'
import * as monaco from 'monaco-editor'

const props = defineProps<{
  modelValue: string
  language: string
  readOnly?: boolean
  diagnostics?: { line: number; column: number; severity: string; message: string }[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: string): void
}>()

const host = ref<HTMLDivElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | null = null
let applying = false

function langOf(name: string) {
  const n = (name || '').toLowerCase()
  if (n === 'go' || n.endsWith('.go')) return 'go'
  if (n.endsWith('.ts') || n.endsWith('.tsx')) return 'typescript'
  if (n.endsWith('.js') || n.endsWith('.jsx')) return 'javascript'
  if (n.endsWith('.json')) return 'json'
  if (n.endsWith('.css')) return 'css'
  if (n.endsWith('.html') || n.endsWith('.vue')) return 'html'
  if (n.endsWith('.md')) return 'markdown'
  if (n.endsWith('.ps1') || n.endsWith('.sh')) return 'shell'
  return 'plaintext'
}

onMounted(() => {
  if (!host.value) return
  editor = monaco.editor.create(host.value, {
    value: props.modelValue || '',
    language: langOf(props.language),
    theme: 'vs-dark',
    automaticLayout: true,
    minimap: { enabled: false },
    fontSize: 12,
    fontFamily: "Consolas, 'Fira Code', monospace",
    readOnly: !!props.readOnly,
    wordWrap: 'on',
    scrollBeyondLastLine: false
  })
  editor.onDidChangeModelContent(() => {
    if (applying || !editor) return
    emit('update:modelValue', editor.getValue())
  })
  applyMarkers()
})

watch(() => props.modelValue, (v) => {
  if (!editor) return
  if (editor.getValue() === v) return
  applying = true
  editor.setValue(v || '')
  applying = false
})

watch(() => props.language, (l) => {
  if (!editor) return
  const model = editor.getModel()
  if (model) monaco.editor.setModelLanguage(model, langOf(l))
})

function applyMarkers() {
  const model = editor?.getModel()
  if (!model) return
  const markers = (props.diagnostics || []).map((d) => ({
    startLineNumber: d.line || 1,
    startColumn: d.column || 1,
    endLineNumber: d.line || 1,
    endColumn: (d.column || 1) + 40,
    message: d.message,
    severity: d.severity === 'WARNING' ? monaco.MarkerSeverity.Warning : monaco.MarkerSeverity.Error
  }))
  monaco.editor.setModelMarkers(model, 'tiancode', markers)
}

watch(() => props.diagnostics, applyMarkers, { deep: true })

onBeforeUnmount(() => {
  editor?.dispose()
  editor = null
})
</script>
