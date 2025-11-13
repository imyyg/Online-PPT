import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import fs from 'fs'
import path from 'path'

// Add dynamic base for GitHub Pages
// 自动从 GitHub Actions 环境变量解析仓库名：owner/repo
// - repo 为 owner.github.io 时，base='/'
// - 其他项目页，base='/${repo}/'
// - 本地/非 Pages 构建，base='/'
const isPages = process.env.GITHUB_PAGES === 'true' || !!process.env.GITHUB_REPOSITORY
const repoSlug = process.env.GITHUB_REPOSITORY || ''
const [owner, repo] = repoSlug.split('/')
const isUserOrOrgSiteRepo = owner && repo && repo.toLowerCase() === `${owner.toLowerCase()}.github.io`
const base = isPages ? (isUserOrOrgSiteRepo ? '/' : (repo ? `/${repo}/` : '/')) : '/'

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

      async function readBody(req) {
        return await new Promise((resolve) => {
          let data = ''
          req.on('data', chunk => { data += chunk })
          req.on('end', () => {
            try { resolve(JSON.parse(data || '{}')) } catch { resolve({}) }
          })
        })
      }

      function resolvePaths(group) {
        const targetGroup = String(group || '').trim() || 'example'
        const slidesDir = path.join(presentationsDir, targetGroup, 'slides')
        const configPath = path.join(presentationsDir, targetGroup, 'slides.config.json')
        if (!fs.existsSync(path.join(presentationsDir, targetGroup))) {
          fs.mkdirSync(path.join(presentationsDir, targetGroup, 'slides'), { recursive: true })
          const blankCfg = {
            title: targetGroup,
            author: '',
            description: '',
            theme: { primaryColor: '#3b82f6', fontFamily: 'system-ui', transition: 'random' },
            settings: { autoPlay: false, autoPlayInterval: 5000, loop: false, showProgress: true, showThumbnails: true, enableKeyboardNav: true, enableTouchNav: true },
            slides: []
          }
          fs.writeFileSync(configPath, JSON.stringify(blankCfg, null, 2))
        }
        return { slidesDir, configPath }
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
            content = content.replace(/<title>.*?<\/title>/, `<title>${title}<\/title>`) 
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

      // Serve presentations files in dev for GET requests
      server.middlewares.use((req, res, next) => {
        if (req.method !== 'GET') return next()
        const url = req.url || ''
        if (url.startsWith('/presentations/')) {
          const rel = decodeURIComponent(url.replace(/^\/presentations\//, ''))
          const filePath = path.join(presentationsDir, rel)
          const normalized = path.normalize(filePath)
          if (normalized.startsWith(presentationsDir) && fs.existsSync(normalized)) {
            const ext = path.extname(normalized).toLowerCase()
            const type = (
              ext === '.json' ? 'application/json' :
              ext === '.html' ? 'text/html; charset=utf-8' :
              ext === '.svg'  ? 'image/svg+xml' :
              ext === '.png'  ? 'image/png' :
              (ext === '.jpg' || ext === '.jpeg') ? 'image/jpeg' :
              'text/plain'
            )
            res.statusCode = 200
            res.setHeader('Content-Type', type)
            try {
              const buf = fs.readFileSync(normalized)
              res.end(buf)
            } catch (e) {
              res.statusCode = 500
              res.end(String(e))
            }
            return
          }
        }
        return next()
      })
    }
  }
}

// Copy presentations folder into dist for static hosting
function copyPresentationsPlugin() {
  return {
    name: 'copy-presentations-plugin',
    apply: 'build',
    closeBundle() {
      try {
        const root = process.cwd()
        const src = path.join(root, 'presentations')
        const dest = path.join(root, 'dist', 'presentations')
        if (fs.existsSync(src)) {
          // Ensure exact sync: remove old dest then copy fresh
          if (fs.existsSync(dest)) {
            fs.rmSync(dest, { recursive: true, force: true })
          }
          fs.mkdirSync(path.join(root, 'dist'), { recursive: true })
          fs.cpSync(src, dest, { recursive: true })
        }
      } catch (e) {
        console.error('[copy-presentations-plugin] Failed to copy presentations:', e)
      }
    }
  }
}

// 生成 404.html 以在 GitHub Pages 上支持 SPA 深链接回退
function spa404Plugin() {
  return {
    name: 'spa-404-plugin',
    apply: 'build',
    closeBundle() {
      try {
        const root = process.cwd()
        const distDir = path.join(root, 'dist')
        const indexPath = path.join(distDir, 'index.html')
        const notFoundPath = path.join(distDir, '404.html')
        if (fs.existsSync(indexPath)) {
          const html = fs.readFileSync(indexPath, 'utf-8')
          fs.writeFileSync(notFoundPath, html, 'utf-8')
        }
      } catch (e) {
        console.error('[spa-404-plugin] Failed to generate 404.html:', e)
      }
    }
  }
}

// https://vite.dev/config/
export default defineConfig({
  base,
  plugins: [vue(), fileOpsPlugin(), copyPresentationsPlugin(), spa404Plugin()],
  server: {
    watch: {
      // Ignore changes inside presentations to prevent dev server full reloads
      ignored: ['**/presentations/**']
    }
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/setup.ts']
  }
})
