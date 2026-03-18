import { motion } from 'motion/react';
import { ArrowRight } from 'lucide-react';

const commands = [
  { name: 'Development Server', description: 'Start the development server', key: '1' },
  { name: 'Start App', description: 'Start the production application', key: '2' },
  { name: 'Database Migration', description: 'Run database migrations', key: '3' },
  { name: 'Docker Up', description: 'Start all Docker containers', key: '4' },
  { name: 'Docker Logs', description: 'View Docker container logs', key: '5' },
];

export function TUIPreview() {
  return (
    <section className="py-24 px-6 relative overflow-hidden">
      {/* Background glow */}
      <div className="absolute inset-0 bg-gradient-to-b from-transparent via-purple-900/10 to-transparent pointer-events-none" />
      
      <div className="max-w-6xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-12"
        >
          <h2 className="text-4xl md:text-5xl mb-4">
            Beautiful terminal interface
          </h2>
          <p className="text-xl text-gray-400 max-w-2xl mx-auto">
            Auto-generated from your config. Clean, intuitive, and fast.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="relative max-w-4xl mx-auto"
        >
          <div className="absolute -inset-1 bg-gradient-to-r from-purple-600 to-pink-600 rounded-2xl opacity-20 blur-2xl" />
          
          {/* Terminal window */}
          <div className="relative bg-[#0d1117] border border-white/10 rounded-2xl overflow-hidden shadow-2xl">
            {/* Terminal header */}
            <div className="flex items-center gap-2 px-6 py-4 border-b border-white/10 bg-white/5">
              <div className="w-3 h-3 rounded-full bg-red-500/70" />
              <div className="w-3 h-3 rounded-full bg-yellow-500/70" />
              <div className="w-3 h-3 rounded-full bg-green-500/70" />
              <span className="ml-4 text-sm text-gray-500 font-mono">nexus</span>
            </div>

            {/* Terminal content */}
            <div className="p-8 font-mono text-sm">
              {/* Header */}
              <div className="mb-6">
                <div className="flex items-center gap-3 mb-2">
                  <div className="text-2xl">⚡</div>
                  <span className="text-xl text-purple-400">Nexus</span>
                  <span className="text-gray-600">v1.0.0</span>
                </div>
                <div className="text-gray-500">Select a command to execute:</div>
              </div>

              {/* Command list */}
              <div className="space-y-2">
                {commands.map((cmd, index) => (
                  <motion.div
                    key={cmd.key}
                    initial={{ opacity: 0, x: -20 }}
                    whileInView={{ opacity: 1, x: 0 }}
                    viewport={{ once: true }}
                    transition={{ duration: 0.3, delay: 0.3 + index * 0.1 }}
                    className={`group flex items-start gap-4 p-4 rounded-lg transition-all cursor-pointer ${
                      index === 0
                        ? 'bg-purple-500/20 border border-purple-500/50'
                        : 'bg-white/5 border border-transparent hover:bg-white/10 hover:border-white/10'
                    }`}
                  >
                    <div className={`flex-shrink-0 w-8 h-8 rounded flex items-center justify-center ${
                      index === 0
                        ? 'bg-purple-500 text-white'
                        : 'bg-white/10 text-gray-500 group-hover:bg-white/20'
                    }`}>
                      {cmd.key}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className={`mb-1 ${index === 0 ? 'text-white' : 'text-gray-300'}`}>
                        {cmd.name}
                      </div>
                      <div className="text-sm text-gray-500">
                        {cmd.description}
                      </div>
                    </div>
                    {index === 0 && (
                      <ArrowRight className="flex-shrink-0 w-5 h-5 text-purple-400" />
                    )}
                  </motion.div>
                ))}
              </div>

              {/* Footer hint */}
              <div className="mt-6 pt-6 border-t border-white/10 text-gray-500 text-xs">
                <div className="flex items-center justify-between">
                  <span>↑↓ Navigate</span>
                  <span>Enter Execute</span>
                  <span>Q Quit</span>
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
