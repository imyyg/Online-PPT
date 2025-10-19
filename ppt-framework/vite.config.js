import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import fs from 'fs'
import path from 'path'

function fileOpsPlugin() {
  return {
    name: 'file-ops-plugin',
    configureServer(server) {
      const root = server.config.root
      const slidesDir = path.join(root, 'slides')
      const presentationsDir = path.join(root, 'presentations')
      const templatesDir = path.join(root, 'public', 'templates')
      const configPath = path.join(root, 'slides.config.json')

      // Ensure presentations directory exists and migrate default example if needed
      try {
        if (!fs.existsSync(presentationsDir)) {
          fs.mkdirSync(presentationsDir, { recursive: true })
        }
        const exampleDir = path.join(presentationsDir, 'example')
        const exampleSlidesDir = path.join(exampleDir, 'slides')
        const exampleConfigPath = path.join(exampleDir, 'slides.config.json')
        if (!fs.existsSync(exampleDir)) {
          fs.mkdirSync(exampleDir, { recursive: true })
          fs.mkdirSync(exampleSlidesDir, { recursive: true })
          // Copy slides.config.json if present
          if (fs.existsSync(configPath)) {
            try {
              fs.copyFileSync(configPath, exampleConfigPath)
            } catch (e) { /* ignore */ }
          } else {
            // Create a blank config
            const blankCfg = {
              title: 'Example Presentation',
              author: '',
              description: '',
              theme: { primaryColor: '#3b82f6', fontFamily: 'system-ui', transition: 'random' },
              settings: { autoPlay: false, autoPlayInterval: 5000, loop: false, showProgress: true, showThumbnails: true, enableKeyboardNav: true, enableTouchNav: true },
              slides: []
            }
            fs.writeFileSync(exampleConfigPath, JSON.stringify(blankCfg, null, 2))
          }
          // Copy slide HTML files
          if (fs.existsSync(slidesDir)) {
            for (const f of fs.readdirSync(slidesDir)) {
              if (f.endsWith('.html')) {
                try {
                  fs.copyFileSync(path.join(slidesDir, f), path.join(exampleSlidesDir, f))
                } catch (e) { /* ignore */ }
              }
            }
          }
        }
      } catch (e) {
        // migration best-effort
      }

      function resolvePaths(group) {
        if (group) {
          const groupDir = path.join(presentationsDir, group)
          const groupSlidesDir = path.join(groupDir, 'slides')
          const groupConfigPath = path.join(groupDir, 'slides.config.json')
          if (!fs.existsSync(groupDir)) fs.mkdirSync(groupDir, { recursive: true })
          if (!fs.existsSync(groupSlidesDir)) fs.mkdirSync(groupSlidesDir, { recursive: true })
          if (!fs.existsSync(groupConfigPath)) {
            const defaultCfg = {
              title: 'New Presentation',
              author: '',
              description: '',
              theme: { primaryColor: '#3b82f6', fontFamily: 'system-ui', transition: 'random' },
              settings: { autoPlay: false, autoPlayInterval: 5000, loop: false, showProgress: true, showThumbnails: true, enableKeyboardNav: true, enableTouchNav: true },
              slides: []
            }
            fs.writeFileSync(groupConfigPath, JSON.stringify(defaultCfg, null, 2))
          }
          return { slidesDir: groupSlidesDir, configPath: groupConfigPath }
        }
        return { slidesDir, configPath }
      }

      function readBody(req) {
        return new Promise((resolve) => {
          const chunks = []
          req.on('data', (c) => chunks.push(c))
          req.on('end', () => {
            try {
              const raw = Buffer.concat(chunks).toString()
              resolve(raw ? JSON.parse(raw) : {})
            } catch (e) {
              resolve({})
            }
          })
        })
      }

      function sendJson(res, data, status = 200) {
        res.statusCode = status
        res.setHeader('Content-Type', 'application/json')
        res.end(JSON.stringify(data))
      }

      server.middlewares.use(async (req, res, next) => {
        if (req.method !== 'POST') return next()

        if (req.url === '/api/slides/create') {
          const body = await readBody(req)
          const filename = String(body.file || '').trim()
          const template = String(body.template || 'blank').trim()
          const title = String(body.title || 'New Slide').trim()
          const group = String(body.group || '').trim()

          if (!filename || !filename.endsWith('.html')) {
            return sendJson(res, { ok: false, error: 'Invalid filename' }, 400)
          }

          const { slidesDir: targetSlidesDir, configPath: targetConfigPath } = resolvePaths(group)
          const newFilePath = path.join(targetSlidesDir, filename)

          try {
            // Load template (fallback to blank)
            let tplPath = path.join(templatesDir, `${template}.html`)
            if (!fs.existsSync(tplPath)) {
              tplPath = path.join(templatesDir, 'blank.html')
            }
            let content = fs.readFileSync(tplPath, 'utf-8')
            // Simple title injection if placeholder exists
            content = content.replace(/<title>.*?<\/title>/, `<title>${title}</title>`) 
            content = content.replace(/<h1[^>]*>.*?<\/h1>/, `<h1>${title}</h1>`) 

            fs.writeFileSync(newFilePath, content, 'utf-8')

            // Update slides.config.json
            const cfg = JSON.parse(fs.readFileSync(targetConfigPath, 'utf-8'))
            const newSlide = {
              id: `slide-${Date.now()}`,
              title,
              file: filename,
              visible: true,
              notes: '',
              duration: null
            }
            cfg.slides.push(newSlide)
            fs.writeFileSync(targetConfigPath, JSON.stringify(cfg, null, 2))

            return sendJson(res, { ok: true, slide: newSlide })
          } catch (e) {
            return sendJson(res, { ok: false, error: e.message || String(e) }, 500)
          }
        }

        if (req.url === '/api/slides/duplicate') {
          const body = await readBody(req)
          const sourceFile = String(body.sourceFile || '').trim()
          const sourceTitle = String(body.sourceTitle || 'Slide').trim()
          const group = String(body.group || '').trim()
          if (!sourceFile || !sourceFile.endsWith('.html')) {
            return sendJson(res, { ok: false, error: 'Invalid source file' }, 400)
          }
          const { slidesDir: targetSlidesDir, configPath: targetConfigPath } = resolvePaths(group)
          const srcPath = path.join(targetSlidesDir, sourceFile)
          if (!fs.existsSync(srcPath)) {
            return sendJson(res, { ok: false, error: 'Source file not found' }, 404)
          }
          const stamp = Date.now()
          const base = sourceFile.replace(/\.html$/, '')
          const newName = `${base}-copy-${stamp}.html`
          const destPath = path.join(targetSlidesDir, newName)

          try {
            const html = fs.readFileSync(srcPath, 'utf-8')
            fs.writeFileSync(destPath, html, 'utf-8')

            const cfg = JSON.parse(fs.readFileSync(targetConfigPath, 'utf-8'))
            const newSlide = {
              id: `slide-${stamp}`,
              title: `${sourceTitle} (Copy)`,
              file: newName,
              visible: true,
              notes: '',
              duration: null
            }
              cfg.slides.push(newSlide)
              fs.writeFileSync(targetConfigPath, JSON.stringify(cfg, null, 2))

            return sendJson(res, { ok: true, slide: newSlide })
          } catch (e) {
            return sendJson(res, { ok: false, error: e.message || String(e) }, 500)
          }
        }

        if (req.url === '/api/slides/reorder') {
          const body = await readBody(req)
          const from = Number(body.from)
          const to = Number(body.to)
          const group = String(body.group || '').trim()
          if (!Number.isInteger(from) || !Number.isInteger(to)) {
            return sendJson(res, { ok: false, error: 'Invalid indices' }, 400)
          }
          try {
            const { configPath: targetConfigPath } = resolvePaths(group)
            const cfg = JSON.parse(fs.readFileSync(targetConfigPath, 'utf-8'))
            const slides = cfg.slides
            if (from < 0 || from >= slides.length || to < 0 || to > slides.length) {
              return sendJson(res, { ok: false, error: 'Index out of range' }, 400)
            }
            const [removed] = slides.splice(from, 1)
            slides.splice(to, 0, removed)
            fs.writeFileSync(targetConfigPath, JSON.stringify(cfg, null, 2))
            return sendJson(res, { ok: true })
          } catch (e) {
            return sendJson(res, { ok: false, error: e.message || String(e) }, 500)
          }
        }

        if (req.url === '/api/slides/delete') {
          const body = await readBody(req)
          let id = String(body.id || '').trim()
          let file = String(body.file || '').trim()
          const group = String(body.group || '').trim()

          try {
            const { slidesDir: targetSlidesDir, configPath: targetConfigPath } = resolvePaths(group)
            const cfg = JSON.parse(fs.readFileSync(targetConfigPath, 'utf-8'))
            // If only id provided, resolve file from config
            if (!file && id) {
              const found = cfg.slides.find(s => s.id === id)
              file = found ? found.file : ''
            }

            // Remove slide entry from config
            cfg.slides = cfg.slides.filter(s => s.id !== id && s.file !== file)
            fs.writeFileSync(targetConfigPath, JSON.stringify(cfg, null, 2))

            // Delete the slide file if present
            if (file) {
              const target = path.join(targetSlidesDir, file)
              if (fs.existsSync(target)) {
                try { fs.unlinkSync(target) } catch (e) { /* ignore unlink errors */ }
              }
            }

            return sendJson(res, { ok: true })
          } catch (e) {
            return sendJson(res, { ok: false, error: e.message || String(e) }, 500)
          }
        }

        if (req.url === '/api/presentations/create') {
          const body = await readBody(req)
          const group = String(body.group || '').trim()
          const title = String(body.title || '').trim()
          const description = String(body.description || '').trim()
          if (!group || /[^a-zA-Z0-9-_]/.test(group)) {
            return sendJson(res, { ok: false, error: 'Invalid group name' }, 400)
          }
          try {
            const { slidesDir: groupSlidesDir, configPath: groupConfigPath } = resolvePaths(group)
            // Create or update config with provided metadata
            let cfg
            if (!fs.existsSync(groupConfigPath)) {
              cfg = {
                title: title || 'New Presentation',
                author: '',
                description: description || '',
                theme: { primaryColor: '#3b82f6', fontFamily: 'system-ui', transition: 'random' },
                settings: { autoPlay: false, autoPlayInterval: 5000, loop: false, showProgress: true, showThumbnails: true, enableKeyboardNav: true, enableTouchNav: true },
                slides: []
              }
            } else {
              try {
                cfg = JSON.parse(fs.readFileSync(groupConfigPath, 'utf-8'))
              } catch (e) {
                cfg = {
                  title: title || 'New Presentation',
                  author: '',
                  description: description || '',
                  theme: { primaryColor: '#3b82f6', fontFamily: 'system-ui', transition: 'random' },
                  settings: { autoPlay: false, autoPlayInterval: 5000, loop: false, showProgress: true, showThumbnails: true, enableKeyboardNav: true, enableTouchNav: true },
                  slides: []
                }
              }
              if (title) cfg.title = title
              if (description) cfg.description = description
            }
            fs.writeFileSync(groupConfigPath, JSON.stringify(cfg, null, 2))
            return sendJson(res, { ok: true, group, paths: { slidesDir: groupSlidesDir, configPath: groupConfigPath }, config: cfg })
          } catch (e) {
            return sendJson(res, { ok: false, error: e.message || String(e) }, 500)
          }
        }
        if (req.url === '/api/config/save') {
          const body = await readBody(req)
          const group = String(body.group || '').trim()
          const cfg = body.config
          if (!group) {
            return sendJson(res, { ok: false, error: 'Missing group' }, 400)
          }
          try {
            const { configPath: targetConfigPath } = resolvePaths(group)
            const finalCfg = (cfg && typeof cfg === 'object') ? cfg : {}
            fs.writeFileSync(targetConfigPath, JSON.stringify(finalCfg, null, 2))
            return sendJson(res, { ok: true })
          } catch (e) {
            return sendJson(res, { ok: false, error: e.message || String(e) }, 500)
          }
        }

        if (req.url === '/api/slides/save') {
          const body = await readBody(req)
          const file = String(body.file || '').trim()
          const group = String(body.group || '').trim()
          const html = String(body.html || '')
          if (!file || !file.endsWith('.html')) {
            return sendJson(res, { ok: false, error: 'Invalid file' }, 400)
          }
          try {
            const { slidesDir: targetSlidesDir } = resolvePaths(group)
            const targetPath = path.join(targetSlidesDir, file)
            fs.writeFileSync(targetPath, html, 'utf-8')
            return sendJson(res, { ok: true })
          } catch (e) {
            return sendJson(res, { ok: false, error: e.message || String(e) }, 500)
          }
        }
        return next()
      })
    }
  }
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), fileOpsPlugin()],
  server: {
    watch: {
      // Ignore changes inside presentations to prevent dev server full reloads
      ignored: ['**/presentations/**']
    }
  }
})
