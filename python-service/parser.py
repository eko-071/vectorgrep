import json
import re
import subprocess
import sys

HEADER_PATTERN = re.compile(r"^[A-Z][A-Z ]*$")
TITLE_PAGE_PATTERN = re.compile(r"\((\w+)\)")
MAX_CHUNK_WORDS = 400

def get_manpage_text(command: str) -> str:
    result = subprocess.run(
        f"man {command} | col -b",
        shell=True,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0 or not result.stdout.strip():
        raise ValueError(f"no man page found for '{command}'")
    return result.stdout

def extract_page_number(first_line: str) -> str:
    match = TITLE_PAGE_PATTERN.search(first_line)
    return match.group(1) if match else "unknown"

def split_into_sections(text: str) -> dict[str, str]:
    """Walk the raw man page line by line, grouping body text under each header."""
    lines = text.splitlines()
    sections: dict[str, list[str]] = {}
    current_section = None

    for line in lines:
        if HEADER_PATTERN.match(line) and line.strip():
            current_section = line.strip()
            sections[current_section] = []
        elif current_section:
            sections[current_section].append(line)

    return {name: "\n".join(body) for name, body in sections.items()}

def clean_section_text(raw: str) -> str:
    """Normalize whitespace artifacts left by col -b: tabs, soft hyphens, wrapped lines."""
    text = raw.replace("\t", " ")
    text = re.sub(r"\u00ad\s*", "", text)  # soft hyphen + the wrap space it leaves behind
    text = re.sub(r" +", " ", text)

    paragraphs = re.split(r"\n\s*\n", text)
    cleaned_paragraphs = []
    for para in paragraphs:
        joined = " ".join(line.strip() for line in para.splitlines() if line.strip())
        if joined:
            cleaned_paragraphs.append(joined)

    return "\n\n".join(cleaned_paragraphs)

def chunk_section(text: str, max_words: int = MAX_CHUNK_WORDS) -> list[str]:
    """Split a section into chunks on paragraph boundaries, merging until near the word limit."""
    paragraphs = text.split("\n\n")
    chunks = []
    current = []
    current_words = 0

    for para in paragraphs:
        para_words = len(para.split())
        if current and current_words + para_words > max_words:
            chunks.append("\n\n".join(current))
            current = []
            current_words = 0
        current.append(para)
        current_words += para_words

    if current:
        chunks.append("\n\n".join(current))

    return chunks

def parse_manpage(command: str) -> list[dict]:
    raw_text = get_manpage_text(command)
    lines = raw_text.splitlines()
    page_number = extract_page_number(lines[0]) if lines else "unknown"

    raw_sections = split_into_sections(raw_text)

    chunks = []
    for section_name, raw_body in raw_sections.items():
        cleaned = clean_section_text(raw_body)
        if not cleaned:
            continue

        for i, chunk_text in enumerate(chunk_section(cleaned)):
            chunks.append({
                "id": f"{command}_{section_name.lower()}_{i}",
                "text": chunk_text,
                "metadata": {
                    "command": command,
                    "section": section_name,
                    "page": page_number,
                },
            })

    return chunks

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("usage: python parser.py <command>")
        sys.exit(1)

    result = parse_manpage(sys.argv[1])
    print(json.dumps(result, indent=2))