import { Download, FileEdit, Play } from 'lucide-react';
import { motion } from 'motion/react';

const steps = [
  {
    icon: Download,
    step: '01',
    title: 'Download & Run',
    description: 'Download the Nexus binary for your platform and execute it. No installation, no dependencies.',
  },
  {
    icon: FileEdit,
    step: '02',
    title: 'Configure Commands',
    description: 'On first run, the setup wizard appears. Define your commands with names, descriptions, and actual commands.',
  },
  {
    icon: Play,
    step: '03',
    title: 'Launch & Execute',
    description: 'Run Nexus anytime. Select from your clean TUI menu and execute commands with a single keystroke.',
  },
];

export function HowItWorks() {
  return (
    <section id="how-it-works" className="py-24 px-6 relative">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl md:text-5xl mb-4">
            How it works
          </h2>
          <p className="text-xl text-gray-400 max-w-2xl mx-auto">
            Get started in minutes, not hours
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-12 relative">
          {/* Connection line for desktop */}
          <div className="hidden md:block absolute top-20 left-0 right-0 h-0.5 bg-gradient-to-r from-transparent via-purple-500/30 to-transparent" />
          
          {steps.map((step, index) => (
            <motion.div
              key={step.step}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.2 }}
              className="relative text-center"
            >
              <div className="relative inline-flex items-center justify-center w-20 h-20 mb-6">
                <div className="absolute inset-0 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 opacity-20 blur-xl" />
                <div className="relative w-16 h-16 rounded-full bg-gradient-to-br from-purple-500/30 to-pink-500/30 border border-purple-500/50 flex items-center justify-center">
                  <step.icon className="w-7 h-7 text-purple-300" />
                </div>
              </div>
              
              <div className="text-sm text-purple-400 mb-2 font-mono">{step.step}</div>
              <h3 className="text-2xl mb-3">{step.title}</h3>
              <p className="text-gray-400 leading-relaxed max-w-sm mx-auto">
                {step.description}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}