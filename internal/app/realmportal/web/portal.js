"use strict";

const FONT_IDS = ["font16-gold", "font16-grey", "font16-blue", "font16-red"];

class BitmapFont {
  constructor(metadata, image) {
    this.metadata = metadata;
    this.image = image;
  }

  measure(text) {
    return Array.from(text).reduce((width, character) => {
      const glyph = this.metadata.glyphs[character] || this.metadata.glyphs["?"];
      return width + (glyph ? glyph.advance : 0);
    }, 0);
  }

  draw(context, text, x, y) {
    for (const character of Array.from(text)) {
      const glyph = this.metadata.glyphs[character] || this.metadata.glyphs["?"];
      if (!glyph) {
        continue;
      }
      context.drawImage(
        this.image,
        glyph.x,
        glyph.y,
        glyph.width,
        glyph.height,
        x + glyph.offset_x,
        y + glyph.offset_y,
        glyph.width,
        glyph.height,
      );
      x += glyph.advance;
    }
  }
}

async function loadFont(id) {
  const response = await fetch(`/account/fonts/${id}.json?revision=3`, { credentials: "same-origin" });
  if (!response.ok) {
    throw new Error(`font metadata unavailable: ${id}`);
  }
  const metadata = await response.json();
  const image = new Image();
  image.src = metadata.image;
  await image.decode();
  return new BitmapFont(metadata, image);
}

function replaceText(element, font) {
  const text = element.textContent.trim();
  const canvas = document.createElement("canvas");
  paintText(canvas, text, font);
  element.appendChild(canvas);
  return { canvas, text };
}

function paintText(canvas, text, font) {
  const width = Math.max(1, font.measure(text));
  canvas.width = width;
  canvas.height = Math.max(1, font.metadata.line_height);
  canvas.className = "d2-text-canvas";
  canvas.setAttribute("aria-hidden", "true");
  font.draw(canvas.getContext("2d"), text, 0, 0);
}

function attachField(field, font) {
  const input = field.querySelector("input");
  const canvas = field.querySelector("canvas");
  const context = canvas.getContext("2d");

  const render = () => {
    const scale = window.devicePixelRatio || 1;
    const bounds = field.getBoundingClientRect();
    canvas.width = Math.max(1, Math.round(bounds.width * scale));
    canvas.height = Math.max(1, Math.round(bounds.height * scale));
    context.setTransform(scale, 0, 0, scale, 0, 0);
    context.clearRect(0, 0, bounds.width, bounds.height);

    let value = input.type === "password" ? "*".repeat(Array.from(input.value).length) : input.value;
    const showCaret = document.activeElement === input && Math.floor(Date.now() / 500) % 2 === 0;
    if (showCaret) {
      value += "_";
    }
    const available = Math.max(1, bounds.width - 28);
    while (value && font.measure(value) > available) {
      value = value.slice(1);
    }
    const y = Math.round((bounds.height - font.metadata.line_height) / 2);
    font.draw(context, value, 14, y);
  };

  input.addEventListener("input", render);
  input.addEventListener("focus", render);
  input.addEventListener("blur", render);
  window.addEventListener("resize", render);
  window.setInterval(render, 250);
  render();
}

async function initialize() {
  try {
    const loaded = await Promise.all(FONT_IDS.map(async (id) => [id, await loadFont(id)]));
    const fonts = new Map(loaded);
    document.querySelectorAll("[data-d2-text]").forEach((element) => {
      const font = fonts.get(element.dataset.font);
      if (font) {
        const rendered = replaceText(element, font);
        const button = element.closest("button");
        if (button && fonts.has("font16-blue")) {
          const normal = () => paintText(rendered.canvas, rendered.text, font);
          const highlighted = () => paintText(rendered.canvas, rendered.text, fonts.get("font16-blue"));
          button.addEventListener("mouseenter", highlighted);
          button.addEventListener("mouseleave", normal);
          button.addEventListener("focus", highlighted);
          button.addEventListener("blur", normal);
        }
      }
    });
    document.querySelectorAll(".bitmap-field").forEach((field) => {
      attachField(field, fonts.get("font16-gold"));
    });
    document.documentElement.classList.add("font-ready");
  } catch (error) {
    console.warn("Realm portal bitmap fonts are unavailable; using accessible fallback text.", error);
  }
}

initialize();
