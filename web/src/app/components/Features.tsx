import { Terminal, Zap, Settings, Box, Shield, Rocket } from 'lucide-react';
import { motion } from 'motion/react';

const features = [
  {
    icon: Terminal,
    title: 'Interactive TUI',
    description: 'Beautiful, auto-generated terminal interface from your YAML config. No manual setup required.',
  },
  {
    icon: Zap,
    title: 'Lightning Fast',
    description: 'Single binary with zero dependencies. Launch instantly and run commands without the overhead.',
  },
  {
    icon: Settings,
    title: 'Simple Configuration',
    description: 'Define commands in clean YAML. The setup wizard guides you through first-time configuration.',
  },
  {
    icon: Box,
    title: 'Cross-Platform',
    description: 'Works seamlessly on Windows, macOS, and Linux. One tool for all your environments.',
  },
  {
    icon: Shield,
    title: 'Open Source',
    description: 'Fully transparent, community-driven development. Contribute, fork, or customize as you need.',
  },
  {
    icon: Rocket,
    title: 'Universal Commands',
    description: 'Run any command: npm scripts, Docker commands, custom scripts, or system utilities.',
  },
];

export function Features() {
  return (
    <section id="features" className="py-24 px-6 relative">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl md:text-5xl mb-4">
            Everything you need
          </h2>
          <p className="text-xl text-gray-400 max-w-2xl mx-auto">
            Nexus brings simplicity and power to your command-line workflow
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
              className="group relative p-8 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 hover:border-purple-500/50 transition-all duration-300"
            >
              <div className="absolute inset-0 bg-gradient-to-br from-purple-500/10 to-pink-500/10 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity" />
              
              <div className="relative">
                <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 border border-purple-500/30 mb-4">
                  <feature.icon className="w-6 h-6 text-purple-400" />
                </div>
                
                <h3 className="text-xl mb-3">{feature.title}</h3>
                <p className="text-gray-400 leading-relaxed">{feature.description}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}