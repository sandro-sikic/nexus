import { motion } from 'motion/react';
import { Copy, Check } from 'lucide-react';
import { useState } from 'react';

const yamlConfig = `title: "My Project Nexus"
ui_mode: group        # show commands organized by group
run_mode: stream      # default: stream output inside the TUI

commands:
  # ── Development ──────────────────────────────────────────
  - name: Dev Server
    description: Start the development server with hot reload
    command: "npm run dev"
    group: Development
    run_mode: handoff   # hand off terminal; Ctrl+C works normally

  - name: Setup and Dev
    description: Install deps then start the dev server
    commands:
      - "npm install"
      - "npm run dev"
    group: Development
    run_mode: handoff

  - name: Watch Tests
    description: Run tests in watch mode
    command: "npm run test:watch"
    group: Development
    run_mode: stream

  # ── Build ────────────────────────────────────────────────
  - name: Build
    description: Production build
    command: "npm run build"
    group: Build

  - name: Preview
    description: Preview the production build locally
    command: "npm run preview"
    group: Build
    run_mode: handoff

  # ── Docker ───────────────────────────────────────────────
  - name: Docker Up
    description: Start all Docker containers
    command: "docker compose up"
    group: Docker
    run_mode: handoff

  - name: Docker Logs
    description: View Docker container logs
    command: "docker compose logs -f"
    group: Docker
    run_mode: stream`;

export function CodeExample() {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(yamlConfig);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section id="documentation" className="py-24 px-6 relative">
      <div className="max-w-6xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-12"
        >
          <h2 className="text-4xl md:text-5xl mb-4">
            Simple configuration
          </h2>
          <p className="text-xl text-gray-400 max-w-2xl mx-auto">
            Define all your commands in one clean YAML file
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="relative"
        >
          <div className="absolute -inset-1 bg-gradient-to-r from-purple-600 to-pink-600 rounded-2xl opacity-20 blur-2xl" />
          
          <div className="relative bg-[#0d1117] border border-white/10 rounded-2xl overflow-hidden">
            {/* Header bar */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-white/10 bg-white/5">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-red-500/70" />
                <div className="w-3 h-3 rounded-full bg-yellow-500/70" />
                <div className="w-3 h-3 rounded-full bg-green-500/70" />
              </div>
              <span className="text-sm text-gray-500 font-mono">nexus.yaml</span>
              <button
                onClick={handleCopy}
                className="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-400 hover:text-white bg-white/5 hover:bg-white/10 rounded-lg transition-all"
              >
                {copied ? (
                  <>
                    <Check className="w-4 h-4" />
                    Copied!
                  </>
                ) : (
                  <>
                    <Copy className="w-4 h-4" />
                    Copy
                  </>
                )}
              </button>
            </div>

            {/* Code content */}
            <div className="p-6 overflow-x-auto">
              <pre className="text-sm md:text-base font-mono leading-relaxed">
                <code>
                  <span className="text-pink-400">title</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">&quot;My Project Nexus&quot;</span>
                  {'\n'}
                  <span className="text-pink-400">ui_mode</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">group</span>
                  <span className="text-gray-500">        # show commands organized by group</span>
                  {'\n'}
                  <span className="text-pink-400">run_mode</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">stream</span>
                  <span className="text-gray-500">      # default: stream output inside the TUI</span>
                  {'\n\n'}
                  <span className="text-purple-400">commands</span>
                  <span className="text-gray-300">:</span>
                  {'\n'}
                  <span className="text-gray-500">  # ── Development ──────────────────────────────────────────</span>
                  {'\n'}
                  <span className="text-pink-400">  - name</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Dev Server</span>
                  {'\n'}
                  <span className="text-pink-400">    description</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Start the development server with hot reload</span>
                  {'\n'}
                  <span className="text-pink-400">    command</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">&quot;npm run dev&quot;</span>
                  {'\n'}
                  <span className="text-pink-400">    group</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">Development</span>
                  {'\n'}
                  <span className="text-pink-400">    run_mode</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">handoff</span>
                  <span className="text-gray-500">   # hand off terminal; Ctrl+C works normally</span>
                  {'\n\n'}
                  <span className="text-pink-400">  - name</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Setup and Dev</span>
                  {'\n'}
                  <span className="text-pink-400">    description</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Install deps then start the dev server</span>
                  {'\n'}
                  <span className="text-pink-400">    commands</span>
                  <span className="text-gray-300">:</span>
                  {'\n'}
                  <span className="text-gray-300">      - </span>
                  <span className="text-green-400">&quot;npm install&quot;</span>
                  {'\n'}
                  <span className="text-gray-300">      - </span>
                  <span className="text-green-400">&quot;npm run dev&quot;</span>
                  {'\n'}
                  <span className="text-pink-400">    group</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">Development</span>
                  {'\n'}
                  <span className="text-pink-400">    run_mode</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">handoff</span>
                  {'\n\n'}
                  <span className="text-pink-400">  - name</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Watch Tests</span>
                  {'\n'}
                  <span className="text-pink-400">    description</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Run tests in watch mode</span>
                  {'\n'}
                  <span className="text-pink-400">    command</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">&quot;npm run test:watch&quot;</span>
                  {'\n'}
                  <span className="text-pink-400">    group</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">Development</span>
                  {'\n'}
                  <span className="text-pink-400">    run_mode</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">stream</span>
                  {'\n\n'}
                  <span className="text-gray-500">  # ── Build ────────────────────────────────────────────────</span>
                  {'\n'}
                  <span className="text-pink-400">  - name</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Build</span>
                  {'\n'}
                  <span className="text-pink-400">    description</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Production build</span>
                  {'\n'}
                  <span className="text-pink-400">    command</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">&quot;npm run build&quot;</span>
                  {'\n'}
                  <span className="text-pink-400">    group</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">Build</span>
                  {'\n\n'}
                  <span className="text-pink-400">  - name</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Preview</span>
                  {'\n'}
                  <span className="text-pink-400">    description</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Preview the production build locally</span>
                  {'\n'}
                  <span className="text-pink-400">    command</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">&quot;npm run preview&quot;</span>
                  {'\n'}
                  <span className="text-pink-400">    group</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">Build</span>
                  {'\n'}
                  <span className="text-pink-400">    run_mode</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">handoff</span>
                  {'\n\n'}
                  <span className="text-gray-500">  # ── Docker ───────────────────────────────────────────────</span>
                  {'\n'}
                  <span className="text-pink-400">  - name</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Docker Up</span>
                  {'\n'}
                  <span className="text-pink-400">    description</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Start all Docker containers</span>
                  {'\n'}
                  <span className="text-pink-400">    command</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">&quot;docker compose up&quot;</span>
                  {'\n'}
                  <span className="text-pink-400">    group</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">Docker</span>
                  {'\n'}
                  <span className="text-pink-400">    run_mode</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">handoff</span>
                  {'\n\n'}
                  <span className="text-pink-400">  - name</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">Docker Logs</span>
                  {'\n'}
                  <span className="text-pink-400">    description</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">View Docker container logs</span>
                  {'\n'}
                  <span className="text-pink-400">    command</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-green-400">&quot;docker compose logs -f&quot;</span>
                  {'\n'}
                  <span className="text-pink-400">    group</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">Docker</span>
                  {'\n'}
                  <span className="text-pink-400">    run_mode</span>
                  <span className="text-gray-300">: </span>
                  <span className="text-blue-400">stream</span>
                </code>
              </pre>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
}