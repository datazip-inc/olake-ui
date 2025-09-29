#!/usr/bin/env node
import fs from 'fs'
import path from 'path'

const SRC_DIR = path.resolve(process.cwd(), 'src')
const VALID_EXTS = ['.ts', '.tsx', '.js', '.jsx', '.mts', '.cts']

const aliasMap = new Map([
  ['api', '@api'],
  ['store', '@store'],
  ['modules', '@modules'],
  ['utils', '@utils'],
  ['types', '@types'],
  ['assets', '@assets'],
])

function listFiles(dir) {
  const out = []
  const stack = [dir]
  while (stack.length) {
    const d = stack.pop()
    const entries = fs.readdirSync(d, { withFileTypes: true })
    for (const e of entries) {
      const p = path.join(d, e.name)
      if (e.isDirectory()) stack.push(p)
      else out.push(p)
    }
  }
  return out
}

function isCodeFile(file) {
  return VALID_EXTS.includes(path.extname(file))
}

function stripKnownExt(p) {
  const ext = path.extname(p)
  if (VALID_EXTS.includes(ext) || ['.json'].includes(ext)) return p.slice(0, -ext.length)
  return p
}

function toAlias(specifier, fileDir) {
  try {
    // Only rewrite ../../ or deeper
    if (!specifier.startsWith('../') || !specifier.includes('../../')) return null

    const absoluteTarget = path.resolve(fileDir, specifier)
    if (!absoluteTarget.startsWith(SRC_DIR)) return null

    // If target points to a directory, try index.ts/tsx/js
    let candidate = absoluteTarget
    if (fs.existsSync(candidate) && fs.statSync(candidate).isDirectory()) {
      const idx = ['index.ts', 'index.tsx', 'index.js', 'index.jsx']
        .map(n => path.join(candidate, n))
        .find(fs.existsSync)
      if (idx) candidate = idx
    } else if (!fs.existsSync(candidate)) {
      // Try appending extensions
      const withExt = ['.ts', '.tsx', '.js', '.jsx'].map(ext => candidate + ext).find(fs.existsSync)
      if (withExt) candidate = withExt
    }

    const relFromSrc = path.relative(SRC_DIR, candidate).replace(/\\/g, '/')
    const withoutExt = stripKnownExt(relFromSrc)
    const [first, ...restParts] = withoutExt.split('/')
    const rest = restParts.join('/')

    const alias = aliasMap.get(first)
    if (alias) {
      return rest ? `${alias}/${rest}` : alias
    }
    // Fallback to root '@' alias
    return `@/${withoutExt}`
  } catch (e) {
    return null
  }
}

function transformFile(file) {
  const original = fs.readFileSync(file, 'utf8')
  let changed = false
  const dir = path.dirname(file)

  const pattern = /\b(from|import)\s*(?:[^'"\n]*?)['"](\.\.(?:\/\.\.)+[^'"\n]*)['"]/g

  const result = original.replace(pattern, (match, kw, spec) => {
    const aliased = toAlias(spec, dir)
    if (aliased) {
      changed = true
      return match.replace(spec, aliased)
    }
    return match
  })

  if (changed) {
    fs.writeFileSync(file, result, 'utf8')
    return true
  }
  return false
}

function main() {
  const files = listFiles(SRC_DIR).filter(isCodeFile)
  let modified = 0
  for (const f of files) {
    if (transformFile(f)) modified++
  }
  console.log(`Alias codemod complete. Modified ${modified} files.`)
}

main()
