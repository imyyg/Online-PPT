# PPT Framework - Web-based Presentation System

A modern web-based presentation framework built with Vue 3, where each slide is an independent HTML file.

## Features

- **HTML-based Slides**: Each slide is a standalone HTML file with full creative freedom
- **Dynamic Loading**: Slides are loaded dynamically with smooth transitions
- **Slide Management**: Intuitive interface for managing slides (add, delete, duplicate, reorder)
- **Presentation Mode**: Full-screen presentation with keyboard/touch navigation
- **Customizable**: Themes, transitions, and settings via configuration
- **No Backend Required**: Pure frontend solution

## Project Structure

```
ppt-framework/
├── presentations/
│   └── <group>/
│       ├── slides/            # Slide HTML files for this group
│       │   ├── slide-1.html
│       │   ├── slide-2.html
│       │   └── slide-3.html
│       └── slides.config.json # Presentation configuration for this group
├── public/
│   └── templates/             # Slide templates
├── src/
│   ├── components/            # Vue components
│   ├── stores/                # Pinia stores
│   └── App.vue                # Main application
```

## Getting Started

### Installation

```bash
npm install
```

### Development

```bash
npm run dev
```

### Build

```bash
npm run build
```

## Usage

### Creating Slides

1. Add HTML files to the `/slides/` directory
2. Each HTML file should be a complete webpage
3. Use the slide manager to organize them

### Example Slide

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My Slide</title>
    <style>
        body {
            height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            font-family: system-ui;
        }
    </style>
</head>
<body>
    <h1>Your Content Here</h1>
</body>
</html>
```

### Configuration

Edit `slides.config.json` to customize:

```json
{
  "title": "My Presentation",
  "theme": {
    "primaryColor": "#3b82f6",
    "transition": "slide"
  },
  "settings": {
    "autoPlay": false,
    "showProgress": true,
    "enableKeyboardNav": true
  },
  "slides": [
    {
      "id": "slide-1",
      "title": "Welcome",
      "file": "slide-1.html",
      "visible": true
    }
  ]
}
```

## Keyboard Shortcuts

- **Arrow Keys**: Navigate slides
- **Space**: Next slide
- **F**: Toggle fullscreen
- **Number Keys**: Jump to slide
- **Escape**: Exit presentation mode

## Templates

Several templates are provided in `/public/templates/`:

- `blank.html`: Empty slide template
- `title.html`: Title slide with gradient background

## Future Enhancements

- Export to PDF
- Cloud storage integration
- Collaborative editing
- More transition effects
- Mobile app companion

## License

MIT