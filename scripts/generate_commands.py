#!/usr/bin/env python3
"""Generate commands.txt from man -k output.

Sections 2, 4, 5, 7, 8: all entries kept (system-level docs).
Section 1 (user commands): filtered to remove GUI apps, ImageMagick,
  Perl internals, niche format converters, and other non-relevant tools.
Section 3: hardcoded curated list of core C library functions.
"""

import subprocess

CURATED_SEC3 = [
    # stdlib
    "abort", "abs", "atexit", "atof", "atoi", "atol", "atoll",
    "bsearch", "calloc", "div", "exit", "free", "getenv", "labs",
    "ldiv", "llabs", "lldiv", "malloc", "mkdtemp", "mkstemp",
    "qsort", "rand", "realloc", "setenv", "srand", "system",
    "unsetenv",
    # string
    "memccpy", "memchr", "memcmp", "memcpy", "memmove", "memset",
    "stpcpy", "stpncpy", "strcasecmp", "strcat", "strchr", "strcmp",
    "strcoll", "strcpy", "strcspn", "strdup", "strerror", "strftime",
    "strlen", "strncasecmp", "strncat", "strncmp", "strncpy",
    "strndup", "strnlen", "strpbrk", "strrchr", "strsignal",
    "strspn", "strstr", "strtod", "strtof", "strtok", "strtol",
    "strtold", "strtoll", "strtoul", "strtoull", "strxfrm",
    # stdio
    "dprintf", "fclose", "feof", "ferror", "fflush", "fgetc",
    "fgets", "fileno", "fopen", "fprintf", "fputc", "fputs",
    "fread", "freopen", "fscanf", "fseek", "fsetpos", "ftell",
    "fwrite", "getc", "getchar", "gets", "perror", "popen",
    "printf", "putc", "putchar", "puts", "remove", "rename",
    "rewind", "scanf", "setbuf", "setvbuf", "snprintf", "sprintf",
    "sscanf", "tmpfile", "tmpnam", "ungetc", "vfprintf", "vprintf",
    "vsnprintf", "vsprintf",
    # math
    "acos", "acosh", "asin", "asinh", "atan", "atan2", "atanh",
    "cbrt", "ceil", "cos", "cosh", "erf", "erfc", "exp", "exp2",
    "expm1", "fabs", "floor", "fma", "fmax", "fmin", "fmod",
    "frexp", "hypot", "ilogb", "ldexp", "lgamma", "llrint",
    "llround", "log", "log10", "log1p", "log2", "logb", "lrint",
    "lround", "modf", "nearbyint", "pow", "remainder", "rint",
    "round", "scalbln", "scalbn", "sin", "sinh", "sqrt", "tan",
    "tanh", "tgamma", "trunc",
    # time
    "asctime", "clock", "ctime", "difftime", "gmtime", "localtime",
    "mktime", "strftime", "time",
    # signal
    "raise", "sigaction", "sigaddset", "sigdelset", "sigemptyset",
    "sigfillset", "sigismember", "signal", "sigpending",
    "sigprocmask", "sigsuspend",
    # wide char
    "wcrtomb", "wcscat", "wcschr", "wcscmp", "wcscoll", "wcscpy",
    "wcscspn", "wcslen", "wcsncat", "wcsncmp", "wcsncpy", "wcspbrk",
    "wcsrchr", "wcsrtombs", "wcsspn", "wcsstr", "wcstod", "wcstof",
    "wcstok", "wcstol", "wcstold", "wcstoll", "wcstombs", "wcstoul",
    "wcstoull", "wcsxfrm", "wctob", "wctomb", "wmemchr", "wmemcmp",
    "wmemcpy", "wmemmove", "wmemset", "wprintf", "wscanf",
    # pthread
    "pthread_cancel", "pthread_cond_broadcast", "pthread_cond_destroy",
    "pthread_cond_init", "pthread_cond_signal", "pthread_cond_timedwait",
    "pthread_cond_wait", "pthread_create", "pthread_detach",
    "pthread_equal", "pthread_exit", "pthread_getspecific",
    "pthread_join", "pthread_key_create", "pthread_key_delete",
    "pthread_kill", "pthread_mutex_destroy", "pthread_mutex_init",
    "pthread_mutex_lock", "pthread_mutex_trylock",
    "pthread_mutex_unlock", "pthread_once", "pthread_rwlock_destroy",
    "pthread_rwlock_init", "pthread_rwlock_rdlock",
    "pthread_rwlock_unlock", "pthread_rwlock_wrlock", "pthread_self",
    "pthread_setcancelstate", "pthread_setcanceltype",
    "pthread_setspecific", "pthread_sigmask",
    # dl
    "dlclose", "dlerror", "dlopen", "dlsym",
    # sockets (section 3 variants — section 2 already covers these too)
    "accept", "bind", "connect", "getaddrinfo", "getnameinfo",
    "getpeername", "getsockname", "getsockopt", "htonl", "htons",
    "inet_addr", "inet_aton", "inet_ntoa", "inet_ntop", "inet_pton",
    "listen", "ntohl", "ntohs", "recv", "recvfrom", "recvmsg",
    "send", "sendmsg", "sendto", "setsockopt", "shutdown",
    "socket", "socketpair",
]

SEC1_BLOCK = {
    # GUI / display servers
    "X", "Xephyr", "Xnest", "Xorg", "Xorg.wrap", "Xserver", "Xvfb",
    "Xwayland", "Xmark", "Hyprland", "Thunar",
    # ImageMagick
    "animate", "compare", "composite", "conjure", "convert", "display",
    "identify", "import", "mogrify", "montage", "stream",
    "Magick++-config", "MagickCore-config", "MagickWand-config",
    "ImageMagick",
    # Perl internals
    "CA.pl", "c2ph", "h2ph", "h2xs", "pl2pm", "splain", "pstruct",
    "pod2man", "pod2text", "pod2html", "pod2usage", "podchecker",
    "podselect",
    # NetPBM format converters
    "411toppm", "any2djvu", "anytopnm", "atobm",
    "giftopnm", "gouldtoppm", "hipstopgm",
    "pamtosrf", "pbmtoascii", "pbmtoepsi", "pbmtoepson",
    "pbmtog3", "pbmtoln03", "pbmtomda", "pbmtoneo",
    "pbmtopgm", "pbmtopi3", "pbmtoplot", "pbmtox10bm",
    "pbmtoxbm", "pbmtoxpm", "pbmtoybm", "pbmtozinc",
    "pcxtoppm", "pgmtofs", "pgmtopbm", "pgmtoppm",
    "pgmtozinc", "pi3toppm", "picttoppm", "pjtoppm",
    "pnmarith", "pnmcat", "pnmcomp", "pnmconvol",
    "pnmenlarge", "pnmfile", "pnmflip", "pnmgamma",
    "pnmhisteq", "pnmhistmap", "pnmindex", "pnminvert",
    "pnmmargin", "pnmmontage", "pnmnlfilt", "pnmnoraw",
    "pnmnorm", "pnmpad", "pnmpaste", "pnmquant",
    "pnmremap", "pnmrotate", "pnmscale", "pnmshear",
    "pnmsmooth", "pnmsplit", "pnmtofits", "pnmtopng",
    "pnmtopnm", "pnmtops", "pnmtorast", "pnmtosgi",
    "pnmtosir", "pnmtotiff", "pnmtoy4m", "ppm3d",
    "ppmbrighten", "ppmchange", "ppmcie", "ppmdim",
    "ppmdist", "ppmdither", "ppmfade", "ppmflash",
    "ppmforge", "ppmhist", "ppmlab", "ppmmake",
    "ppmmix", "ppmnorm", "ppmntsc", "ppmpat",
    "ppmquant", "ppmquantall", "ppmqvga", "ppmrainbow",
    "ppmrelief", "ppmshadow", "ppmshift", "ppmspread",
    "ppmtoacad", "ppmtobmp", "ppmtoeyuv", "ppmtogif",
    "ppmtoicr", "ppmtoilbm", "ppmtojpeg", "ppmtoleaf",
    "ppmtolj", "ppmtomitsu", "ppmtoneo", "ppmtopcx",
    "ppmtopgm", "ppmtopict", "ppmtopj", "ppmtopjxl",
    "ppmtopuzz", "ppmtorgb3", "ppmtosixel", "ppmtotga",
    "ppmtoxpm", "ppmtoyuv", "ppmtoyuvsplit",
    "qrttoppm", "rasttopnm", "rawtopgm", "rawtoppm",
    "rgb3toppm", "rpxtoppm",
    "sirtopnm", "sldtoppm", "spctoppm", "spottopgm",
    "tgatoppm", "tifftopnm",
    "xbmtopbm", "xbmtoxpm", "xpmtoppm",
    "y4mtopnm", "yuvsplittoppm", "zeisstopnm",
    # Audio-only tools
    "aafire", "aconnect", "alsabat", "alsactl", "alsaloop",
    "alsamixer", "alsatplg", "alsaucm", "amidi", "amixer",
    "aplay", "aplaymidi", "aserver",
    "esd", "esd-config", "esdcat", "esdctl", "esdloop",
    "esdplay", "esdrec", "esdsend", "esdmon",
    "mpg321", "mpg123", "timidity",
    # Math/stats languages
    "R", "Rscript", "Singular", "TSingular", "ESingular",
    # Obsolete / very niche
    "FileCheck", "ctags", "ctangle", "ctie",
    "dbiprof", "dbiproxy", "dbilogstrip", "dbish",
    "dwp", "elfedit",
    "freeciv", "freeciv-gtk3", "freeciv-gtk4",
    "freeciv-qt", "freeciv-xaw", "freeciv-client",
    "freeciv-server", "freeciv-mp-gtk3", "freeciv-mp-qt",
    "freeciv-mp-trident", "freeciv-ruledit",
    "geomyid", "getkeycodes",
    "lbrowse",
    "mscompress", "msfilter", "mshowfat", "mussh",
    "netdiscover", "netperf", "netperf_bas", "netperf_cpu",
    "netperf_ns", "nns", "o2", "o2registry",
    "pf2afm", "pfbtops", "pk2bm", "ps2epsi", "ps2frag",
    "ps2pkm", "ps2pk", "ps2ps2",
}

SEC1_BLOCK_PREFIXES = ["perl", "pod2"]


def get_man_entries(section: str | int) -> list[str]:
    result = subprocess.run(
        ["man", "-k", "-s", str(section), "."],
        capture_output=True, text=True, timeout=30,
    )
    if result.returncode != 0:
        return []
    names = set()
    for line in result.stdout.splitlines():
        name = line.split()[0].strip()
        if name and not name.startswith("-"):
            names.add(name)
    return sorted(names)


def is_blocked_sec1(name: str) -> bool:
    return name in SEC1_BLOCK or any(
        name.startswith(p) for p in SEC1_BLOCK_PREFIXES
    )


def main():
    SECTIONS = [
        (2, "System Calls"),
        (4, "Devices"),
        (5, "File Formats"),
        (7, "Miscellaneous"),
        (8, "Admin Commands"),
    ]
    HEADER = "# ============================================================================"

    sec1_entries = sorted(set(get_man_entries(1)) | set(get_man_entries("1p")))
    kept_sec1 = [n for n in sec1_entries if not is_blocked_sec1(n)]
    section_map = {num: get_man_entries(num) for num, _ in SECTIONS}

    lines = [HEADER, "# Auto-generated commands.txt", HEADER, ""]
    for num, title in SECTIONS:
        lines += [HEADER, f"# Section {num}: {title}", HEADER, ""]
        lines += section_map[num] + [""]
    lines += [HEADER, "# Section 1: User Commands (filtered)", HEADER, ""]
    lines += kept_sec1 + [""]
    lines += [HEADER, "# Section 3: Library Functions (curated)", HEADER, ""]
    lines += CURATED_SEC3 + [""]

    path = "go-api/commands.txt"
    with open(path, "w") as f:
        f.write("\n".join(lines) + "\n")

    blocked = len(sec1_entries) - len(kept_sec1)
    total = sum(len(v) for v in section_map.values()) + len(kept_sec1) + len(CURATED_SEC3)
    print(f"Section 1: {len(kept_sec1)} kept, {blocked} blocked out of {len(sec1_entries)} total")
    for num, _ in SECTIONS:
        print(f"Section {num}: {len(section_map[num])}")
    print(f"Section 3: {len(CURATED_SEC3)} (curated)")
    print(f"Total: {total}")
    print(f"Written to {path}")


if __name__ == "__main__":
    main()
