import { readdirSync, readFileSync, statSync } from 'node:fs';
import { extname, join, resolve } from 'node:path';

const DIST_DIR = resolve(import.meta.dirname, '../dist');

const FORBIDDEN_PATTERNS = [
  'MOCK-0001',
  'com.example.mock.',
  'MockBridge',
  'ZodObject',
  'ZodError',
];

function getAllFiles(dir) {
  let results = [];
  const list = readdirSync(dir);
  for (const file of list) {
    const fullPath = join(dir, file);
    const stat = statSync(fullPath);
    if (stat.isDirectory()) {
      results = results.concat(getAllFiles(fullPath));
    } else {
      results.push(fullPath);
    }
  }
  return results;
}

try {
  const files = getAllFiles(DIST_DIR).filter((f) => extname(f) === '.js');
  if (files.length === 0) {
    console.error(`[check:no-mock] FAILED: No JavaScript bundles found in ${DIST_DIR}`);
    process.exit(1);
  }

  let violations = 0;
  for (const filePath of files) {
    const content = readFileSync(filePath, 'utf-8');
    for (const pattern of FORBIDDEN_PATTERNS) {
      if (content.includes(pattern)) {
        console.error(
          `[check:no-mock] VIOLATION: Found forbidden marker "${pattern}" in bundle ${filePath}`
        );
        violations++;
      }
    }
  }

  if (violations > 0) {
    console.error(
      `[check:no-mock] FAILED: ${violations} forbidden mock/zod markers detected in production bundle.`
    );
    process.exit(1);
  }

  console.log(
    `[check:no-mock] PASSED: 0 mock markers or runtime schema artifacts detected across ${files.length} production bundle(s).`
  );
  process.exit(0);
} catch (err) {
  console.error(`[check:no-mock] ERROR: ${err.message}`);
  process.exit(1);
}
