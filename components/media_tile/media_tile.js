import { LitElement, css, html } from "https://esm.sh/lit";
import { SvgIcon } from "../svg_icon/svg_icon.js";

class MediaTile extends LitElement {
  static styles = css`
    :host {
      display: block;
      position: relative;
      border-radius: 12px;
      overflow: hidden;
      cursor: pointer;
      border: 1px solid var(--mesop-outline-variant-color);
      height: 250px;
      /* Add a subtle background for definition on all themes */
      background-color: rgba(255, 255, 255, 0.05);
    }

    .preview {
      width: 100%;
      height: 100%;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .preview img,
    .preview video {
      width: 100%;
      height: 100%;
      object-fit: cover; /* Use cover to fill the tile, looks less square */
    }

    .preview .icon {
      width: 64px;
      height: 64px;
      color: var(--mesop-on-surface-variant-color);
    }

    .overlay {
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      /* Add a gradient scrim for text readability on hover */
      background: linear-gradient(
        to top,
        rgba(0, 0, 0, 0.7) 0%,
        rgba(0, 0, 0, 0.3) 50%,
        transparent 100%
      );
      display: flex;
      flex-direction: column;
      justify-content: flex-end;
      padding: 12px;
      pointer-events: none; /* Allow clicks to pass through the overlay */
      opacity: 0; /* Hide the entire overlay by default */
      transition: opacity 0.2s ease-in-out;
    }

    .overlay > * {
      pointer-events: auto; /* The content itself is clickable */
    }

    :host(:hover) .overlay {
      opacity: 1;
    }

    .pills-container {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
    }

    .pill {
      background-color: rgba(255, 255, 255, 0.8);
      color: #202124;
      padding: 4px 8px;
      border-radius: 6px;
      font-size: 12px;
      font-weight: 500;
    }

    audio {
      width: 100%;
      margin-bottom: 8px;
      filter: invert(1) grayscale(1) contrast(1.5);
    }
  `;

  static get properties() {
    return {
      mediaType: { type: String },
      thumbnailSrc: { type: String },
      audioSrc: { type: String },
      pillsJson: { type: String },
    };
  }

  constructor() {
    super();
    this.mediaType = "";
    this.thumbnailSrc = "";
    this.audioSrc = "";
    this.pillsJson = "[]";
    this.addEventListener("click", this.handleClick);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this.removeEventListener("click", this.handleClick);
  }

  handleClick(e) {
    const event = new CustomEvent("clickEvent", {
      bubbles: true,
      composed: true,
    });
    this.dispatchEvent(event);
  }

  renderPreview() {
    switch (this.mediaType) {
      case "image":
        return html`<img .src=${this.thumbnailSrc} />`;
      case "video":
        return html`<video
          .src=${this.thumbnailSrc}
          muted
          autoplay
          loop
          playsinline
        ></video>`;
      case "audio":
        return html`<div class="icon"><svg-icon .iconName=${'music_note'}></svg-icon></div>`;
      default:
        return html`<div class="icon"><svg-icon .iconName=${'help'}></svg-icon></div>`;
    }
  }

  renderPills() {
    try {
      const pills = JSON.parse(this.pillsJson);
      return pills.map((pill) => html`<div class="pill">${pill.label}</div>`);
    } catch (e) {
      console.error("Error parsing pillsJson:", e);
      return html``;
    }
  }

  render() {
    return html`
      <div class="preview">${this.renderPreview()}</div>
      <div class="overlay">
        ${this.mediaType === "audio"
          ? html`<audio controls .src=${this.audioSrc}></audio>`
          : ""}
        <div class="pills-container">${this.renderPills()}</div>
      </div>
    `;
  }
}

customElements.define("media-tile", MediaTile);
