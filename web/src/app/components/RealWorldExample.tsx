import { motion } from 'motion/react';
import { ArrowRight, Terminal } from 'lucide-react';

const scenarios = [
  {
    title: 'Development Workflow',
    before: [
      'Search through terminal history',
      'Remember which port the dev server uses',
      'Type out docker compose up -d every time',
      'Forget migration commands',
    ],
    after: [
      'Open Nexus',
      'Press 1 for dev server',
      'Press 4 for Docker',
      'Done in seconds',
    ],
  },
  {
    title: 'DevOps Tasks',
    before: [
      'Check documentation for deployment commands',
      'Copy-paste from notes',
      'Hope you got the right environment variables',
      'Cross fingers',
    ],
    after: [
      'Run Nexus',
      'Select deployment command',
      'Execute with confidence',
      'Consistent every time',
    ],
  },
];

export function RealWorldExample() {
  return (
    <section id="examples" className="py-24 px-6 relative">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl md:text-5xl mb-4">
            Turn chaos into clarity
          </h2>
          <p className="text-xl text-gray-400 max-w-2xl mx-auto">
            See how Nexus transforms your daily workflow
          </p>
        </motion.div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {scenarios.map((scenario, index) => (
            <motion.div
              key={scenario.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.2 }}
              className="relative"
            >
              <div className="absolute -inset-1 bg-gradient-to-r from-purple-600/20 to-pink-600/20 rounded-2xl blur-xl" />
              
              <div className="relative bg-white/5 border border-white/10 rounded-2xl p-8">
                <div className="flex items-center gap-3 mb-6">
                  <Terminal className="w-6 h-6 text-purple-400" />
                  <h3 className="text-2xl">{scenario.title}</h3>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                  {/* Before */}
                  <div>
                    <div className="text-sm text-red-400 mb-4 font-mono">BEFORE</div>
                    <div className="space-y-3">
                      {scenario.before.map((step, i) => (
                        <div key={i} className="flex items-start gap-3">
                          <div className="flex-shrink-0 w-6 h-6 rounded-full bg-red-500/20 border border-red-500/30 flex items-center justify-center text-xs text-red-400">
                            {i + 1}
                          </div>
                          <p className="text-gray-400 text-sm leading-relaxed">{step}</p>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Arrow */}
                  

                  {/* After */}
                  <div>
                    <div className="text-sm text-green-400 mb-4 font-mono">WITH NEXUS</div>
                    <div className="space-y-3">
                      {scenario.after.map((step, i) => (
                        <div key={i} className="flex items-start gap-3">
                          <div className="flex-shrink-0 w-6 h-6 rounded-full bg-green-500/20 border border-green-500/30 flex items-center justify-center text-xs text-green-400">
                            {i + 1}
                          </div>
                          <p className="text-gray-300 text-sm leading-relaxed">{step}</p>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}