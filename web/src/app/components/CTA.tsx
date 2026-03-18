import { motion } from 'motion/react';
import { Download, Github, ArrowRight } from 'lucide-react';

export function CTA() {
  return (
    <section className="py-24 px-6 relative overflow-hidden">
      {/* Background effects */}
      <div className="absolute inset-0 bg-gradient-to-b from-purple-900/20 via-pink-900/20 to-purple-900/20 pointer-events-none" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-purple-600/20 via-transparent to-transparent pointer-events-none" />
      
      <div className="max-w-5xl mx-auto relative z-10">
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="relative"
        >
          <div className="absolute -inset-1 bg-gradient-to-r from-purple-600 to-pink-600 rounded-3xl opacity-20 blur-2xl" />
          
          <div className="relative bg-gradient-to-br from-white/10 to-white/5 border border-white/20 rounded-3xl p-12 md:p-16 text-center backdrop-blur-sm">
            <h2 className="text-4xl md:text-5xl lg:text-6xl mb-6">
              Ready to simplify your
              <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-pink-400">
                terminal workflow?
              </span>
            </h2>
            
            <p className="text-xl text-gray-300 mb-10 max-w-2xl mx-auto">
              Join developers worldwide who've stopped memorizing commands and started shipping faster with Nexus.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-8">
              <button className="group relative inline-flex items-center gap-2 px-8 py-4 bg-gradient-to-r from-purple-600 to-pink-600 rounded-lg overflow-hidden transition-all hover:scale-105 shadow-lg shadow-purple-500/25">
                <div className="absolute inset-0 bg-gradient-to-r from-purple-500 to-pink-500 opacity-0 group-hover:opacity-100 transition-opacity" />
                <Download className="w-5 h-5 relative z-10" />
                <span className="text-lg relative z-10">Download Now</span>
                <ArrowRight className="w-5 h-5 relative z-10 group-hover:translate-x-1 transition-transform" />
              </button>
              
              <button className="inline-flex items-center gap-2 px-8 py-4 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg transition-all hover:border-white/20">
                <Github className="w-5 h-5" />
                <span className="text-lg">Star on GitHub</span>
              </button>
            </div>

            <div className="flex flex-wrap justify-center gap-6 text-sm text-gray-400">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-green-400" />
                <span>Open Source</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-purple-400" />
                <span>MIT License</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-pink-400" />
                <span>Cross-Platform</span>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
