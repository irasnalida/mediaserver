import "movi-player";

const state = {
  library: null,
  dir: "",
};

// Track playable files in current directory & currently active index
let currentPlayableEntries = [];
let activePlayableIndex = -1;

const DESKTOP_QUERY = window.matchMedia("(min-width: 800px)");

const libSelect = document.getElementById("librarySelect");
const breadcrumbEl = document.getElementById("breadcrumb");
const entryListEl = document.getElementById("entryList");
const nowPlayingTitleEl = document.getElementById("nowPlayingTitle");
const nowPlayingMetaEl = document.getElementById("nowPlayingMeta");
const videoEl = document.getElementById("videoPlayer");
const audioEl = document.getElementById("audioPlayer");
const imageEl = document.getElementById("imagePlayer");
const imageModal = document.getElementById("imageModal");
const imageModalImg = document.getElementById("imageModalImg");
const imageModalClose = document.getElementById("imageModalClose");
const refreshBtn = document.getElementById("refreshBtn");

function makeUrl(lib, dir) {
  const p = new URLSearchParams();
  if (lib) p.set("lib", lib);
  if (dir) p.set("dir", dir);
  const q = p.toString();
  return q ? `?${q}` : window.location.pathname;
}

async function navigateTo(library, dir) {
  state.library = library;
  state.dir = dir;
  libSelect.value = library;
  history.pushState({ lib: library, dir: dir }, "", makeUrl(library, dir));
  await browse();
}

async function init() {
  let names = [];
  try {
    const res = await fetch("/api/libraries");
    names = await res.json();
  } catch (err) {
    entryListEl.innerHTML = '<li class="empty">Could not reach the server.</li>';
    return;
  }

  if (!names || names.length === 0) {
    entryListEl.innerHTML =
      '<li class="empty">No libraries configured. Edit config.json with a folder path and restart the server.</li>';
    libSelect.innerHTML = "";
    return;
  }

  libSelect.innerHTML = names
    .map((n) => `<option value="${escapeHtml(n)}">${escapeHtml(n)}</option>`)
    .join("");

  const urlParams = new URLSearchParams(window.location.search);
  const requestedLib = urlParams.get("lib");
  const requestedDir = urlParams.get("dir") || "";

  if (requestedLib && names.includes(requestedLib)) {
    state.library = requestedLib;
    state.dir = requestedDir;
    libSelect.value = requestedLib;
  } else {
    state.library = names[0];
    state.dir = "";
  }

  history.replaceState({ lib: state.library, dir: state.dir }, "", makeUrl(state.library, state.dir));
  await browse();
}

libSelect.addEventListener("change", () => {
  navigateTo(libSelect.value, "");
});

refreshBtn.addEventListener("click", browse);

window.addEventListener("popstate", async (e) => {
  if (e.state) {
    state.library = e.state.lib || libSelect.value;
    state.dir = e.state.dir || "";
  } else {
    const p = new URLSearchParams(window.location.search);
    state.library = p.get("lib") || libSelect.value;
    state.dir = p.get("dir") || "";
  }
  libSelect.value = state.library;
  await browse();
});

async function browse() {
  if (!state.library) return;
  entryListEl.innerHTML = '<li class="empty">Loading…</li>';

  const params = new URLSearchParams({ lib: state.library, dir: state.dir });
  let data;
  try {
    const res = await fetch(`/api/browse?${params.toString()}`);
    if (!res.ok) throw new Error("bad response");
    data = await res.json();
  } catch (err) {
    entryListEl.innerHTML = '<li class="empty">Could not load this folder.</li>';
    return;
  }

  state.dir = data.dir || "";
  renderBreadcrumb();
  renderEntries(data.entries || []);
}

function renderBreadcrumb() {
  const parts = state.dir ? state.dir.split("/") : [];
  const crumbs = [{ label: state.library, dir: "" }];
  let acc = "";
  parts.forEach((p) => {
    acc = acc ? `${acc}/${p}` : p;
    crumbs.push({ label: p, dir: acc });
  });

  breadcrumbEl.innerHTML = crumbs
    .map((c, i) => {
      const isLast = i === crumbs.length - 1;
      return `<span class="crumb${isLast ? " current" : ""}" data-dir="${escapeHtml(
        c.dir
      )}">${escapeHtml(c.label)}</span>`;
    })
    .join('<span class="sep">/</span>');

  breadcrumbEl.querySelectorAll(".crumb:not(.current)").forEach((el) => {
    el.addEventListener("click", () => {
      navigateTo(state.library, el.dataset.dir);
    });
  });
}

function renderEntries(entries) {
  currentPlayableEntries = [];
  activePlayableIndex = -1;

  if (entries.length === 0) {
    entryListEl.innerHTML = '<li class="empty">This folder is empty.</li>';
    return;
  }

  entryListEl.innerHTML = "";
  entries.forEach((entry) => {
    const li = document.createElement("li");

    const icon = document.createElement("span");
    icon.className = "icon";
    icon.textContent = iconFor(entry.type);

    const name = document.createElement("span");
    name.className = "entry-name";
    name.textContent = entry.name;

    li.appendChild(icon);
    li.appendChild(name);

    if (entry.type === "folder") {
      li.addEventListener("click", () => {
        navigateTo(state.library, entry.path);
      });
    } else {
      const size = document.createElement("span");
      size.className = "file-size";
      size.textContent = formatSize(entry.size);
      li.appendChild(size);

      // Track index of playable media files
      const playableIndex = currentPlayableEntries.length;
      currentPlayableEntries.push({ entry, el: li });

      li.addEventListener("click", () => openByIndex(playableIndex));
    }

    entryListEl.appendChild(li);
  });
}

function openByIndex(index) {
  if (index < 0 || index >= currentPlayableEntries.length) return;
  activePlayableIndex = index;
  const { entry, el } = currentPlayableEntries[index];
  open(entry, el);
}

function navigateStep(direction) {
  if (currentPlayableEntries.length === 0) return;

  // If nothing is open yet, start from the first or last item
  if (activePlayableIndex === -1) {
    openByIndex(direction > 0 ? 0 : currentPlayableEntries.length - 1);
    return;
  }

  const nextIndex = activePlayableIndex + direction;
  if (nextIndex >= 0 && nextIndex < currentPlayableEntries.length) {
    openByIndex(nextIndex);
  }
}

function iconFor(type) {
  switch (type) {
    case "folder":
      return "📁";
    case "video":
      return "🎬";
    case "audio":
      return "🎵";
    case "image":
      return "🖼️";
    default:
      return "📄";
  }
}

function open(entry, listItemEl) {
  document
    .querySelectorAll("#entryList li")
    .forEach((el) => el.classList.remove("playing"));
  listItemEl.classList.add("playing");
  listItemEl.scrollIntoView({ block: "nearest", behavior: "smooth" });

  nowPlayingTitleEl.textContent = entry.name;
  nowPlayingMetaEl.textContent = `${entry.type} · ${formatSize(entry.size)}`;

  // Pause active players safely
  if (typeof videoEl.pause === "function") videoEl.pause();
  videoEl.classList.remove("active");

  if (typeof audioEl.pause === "function") audioEl.pause();
  audioEl.classList.remove("active");
  imageEl.classList.remove("active");

  if (entry.type === "video") {
    closeImageModal();
    audioEl.removeAttribute("src");
    audioEl.load();

    videoEl.src = entry.url;
    videoEl.classList.add("active");
    if (typeof videoEl.play === "function") {
      videoEl.play().catch((err) => {
        if (err.name !== "AbortError") console.error("Playback error:", err);
      });
    }
  } else if (entry.type === "audio") {
    closeImageModal();
    videoEl.removeAttribute("src");
    if (typeof videoEl.load === "function") videoEl.load();

    audioEl.src = entry.url;
    audioEl.classList.add("active");
    audioEl.play().catch((err) => {
      if (err.name !== "AbortError") console.error("Audio playback error:", err);
    });
  } else if (entry.type === "image") {
    videoEl.removeAttribute("src");
    if (typeof videoEl.load === "function") videoEl.load();

    if (DESKTOP_QUERY.matches) {
      closeImageModal();
      imageEl.src = entry.url;
      imageEl.classList.add("active");
    } else {
      openImageModal(entry.url);
    }
  }
}

function openImageModal(url) {
  imageModalImg.src = url;
  imageModal.classList.add("open");
}

function closeImageModal() {
  imageModal.classList.remove("open");
  imageModalImg.src = "";
}

imageModalClose.addEventListener("click", closeImageModal);
imageModal.addEventListener("click", (e) => {
  if (e.target === imageModal) closeImageModal();
});

// Global Keyboard Navigation
document.addEventListener("keydown", (e) => {
  const activeTag = document.activeElement?.tagName;
  if (["INPUT", "TEXTAREA", "SELECT"].includes(activeTag)) return;

  if (e.key === "Escape") {
    closeImageModal();
    return;
  }

  if (e.key === "ArrowDown" || e.key === "ArrowRight") {
    e.preventDefault();
    navigateStep(1);
  } else if (e.key === "ArrowUp" || e.key === "ArrowLeft") {
    e.preventDefault();
    navigateStep(-1);
  }
});

function formatSize(bytes) {
  if (bytes === undefined || bytes === null) return "";
  const units = ["B", "KB", "MB", "GB"];
  let size = bytes;
  let i = 0;
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024;
    i++;
  }
  return `${size.toFixed(1)} ${units[i]}`;
}

function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}

init();
