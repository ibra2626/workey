// PreToolUse guard: blocks force push, hook bypass and an oversized CURRENT snapshot.
// Exit code 2 blocks the tool call and returns stderr to Claude.
import fs from 'node:fs';
import path from 'node:path';

const CURRENT_LIMIT = 8000;

function block(message) {
  process.stderr.write(`${message}\n`);
  process.exit(2);
}

function readStdin() {
  try {
    return JSON.parse(fs.readFileSync(0, 'utf8') || '{}');
  } catch {
    return {};
  }
}

function checkBash(command) {
  for (const segment of command.split(/&&|\|\||;|\|/)) {
    if (/\bgit\b/.test(segment) && /\bpush\b/.test(segment)
      && (/\s--force(-with-lease|-if-includes)?\b/.test(segment) || /\s-[a-zA-Z]*f[a-zA-Z]*\b/.test(segment) || /\s\+\S/.test(segment))) {
      block('Force push is not allowed in this project (AGENTS.md). Push without force or ask the user.');
    }
    if (/\bgit\b/.test(segment) && /\s--no-verify\b/.test(segment)) {
      block('Bypassing Git hooks (--no-verify) is not allowed in this project (AGENTS.md).');
    }
  }
}

function applyEdit(text, edit) {
  if (typeof edit.old_string !== 'string' || typeof edit.new_string !== 'string') return text;
  return edit.replace_all ? text.split(edit.old_string).join(edit.new_string) : text.replace(edit.old_string, () => edit.new_string);
}

function checkCurrent(tool, input, projectDir) {
  const file = input.file_path;
  if (typeof file !== 'string') return;
  const current = path.resolve(projectDir, '.agent/state/CURRENT.md');
  if (path.resolve(projectDir, file).toLowerCase() !== current.toLowerCase()) return;

  let next;
  if (tool === 'Write') {
    next = input.content ?? '';
  } else {
    let existing = '';
    try { existing = fs.readFileSync(current, 'utf8'); } catch { return; }
    const edits = tool === 'MultiEdit' ? (input.edits ?? []) : [input];
    next = edits.reduce(applyEdit, existing);
  }
  const length = [...next].length;
  if (length > CURRENT_LIMIT) {
    block(`CURRENT.md would be ${length} characters (limit ${CURRENT_LIMIT}). Summarize it as a replacement snapshot; keep constraints, active paths, checks, blockers and next step.`);
  }
}

const event = readStdin();
const tool = event.tool_name;
const input = event.tool_input ?? {};
const projectDir = process.env.CLAUDE_PROJECT_DIR || event.cwd || process.cwd();

if (tool === 'Bash' && typeof input.command === 'string') checkBash(input.command);
if (['Write', 'Edit', 'MultiEdit'].includes(tool)) checkCurrent(tool, input, projectDir);
process.exit(0);
