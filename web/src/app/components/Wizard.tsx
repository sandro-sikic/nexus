import { Wand2, FileQuestion, Terminal, CheckCircle2, ArrowRight } from 'lucide-react';

export function Wizard() {
  const wizardSteps = [
    {
      title: 'Project Details',
      description: 'Name your configuration and set basic preferences',
      icon: Terminal,
    },
    {
      title: 'Command Groups',
      description: 'Organize commands into logical categories',
      icon: CheckCircle2,
    },
    {
      title: 'Run Modes',
      description: 'Choose execution behavior for each command',
      icon: ArrowRight,
    },
  ];

  return (
    <section id="wizard" className="py-24 px-6 relative">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="text-center mb-16">
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-purple-500/10 border border-purple-500/20 mb-6">
            <Wand2 className="w-4 h-4 text-purple-400" />
            <span className="text-sm text-purple-300">Interactive Setup</span>
          </div>
          <h2 className="text-4xl md:text-5xl mb-6">
            Guided Configuration Wizard
          </h2>
          <p className="text-xl text-gray-400 max-w-3xl mx-auto">
            Not sure where to start? Let the interactive wizard guide you through setting up your perfect configuration.
          </p>
        </div>

        {/* When to use wizard */}
        <div className="grid md:grid-cols-2 gap-8 mb-16">
          <div className="bg-white/5 border border-white/10 rounded-2xl p-8 hover:border-purple-500/30 transition-all">
            <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center mb-4">
              <Terminal className="w-6 h-6" />
            </div>
            <h3 className="text-2xl mb-3">Launch with Flag</h3>
            <p className="text-gray-400 mb-4">
              Start the wizard anytime by running Nexus with the wizard flag:
            </p>
            <div className="bg-black/40 rounded-lg p-4 font-mono text-sm border border-white/10">
              <span className="text-gray-500">$</span> <span className="text-purple-400">nexus</span> <span className="text-pink-400">--wizard</span>
            </div>
          </div>

          <div className="bg-white/5 border border-white/10 rounded-2xl p-8 hover:border-purple-500/30 transition-all">
            <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center mb-4">
              <FileQuestion className="w-6 h-6" />
            </div>
            <h3 className="text-2xl mb-3">Auto-Launch</h3>
            <p className="text-gray-400 mb-4">
              The wizard automatically starts when no configuration file is found:
            </p>
            <div className="bg-black/40 rounded-lg p-4 font-mono text-sm border border-white/10">
              <span className="text-gray-500">$</span> <span className="text-purple-400">nexus</span>
              <div className="text-yellow-400 text-xs mt-2">⚠ No config found. Starting wizard...</div>
            </div>
          </div>
        </div>

        {/* Wizard Flow */}
        <div className="bg-gradient-to-br from-white/5 to-white/0 border border-white/10 rounded-2xl p-8 md:p-12">
          <h3 className="text-2xl mb-8 text-center">Step-by-Step Configuration</h3>
          
          <div className="grid md:grid-cols-3 gap-6">
            {wizardSteps.map((step, index) => (
              <div key={index} className="relative">
                {/* Step Card */}
                <div className="bg-black/40 border border-white/10 rounded-xl p-6 hover:border-purple-500/30 transition-all h-full">
                  <div className="flex items-start gap-4 mb-4">
                    <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center flex-shrink-0">
                      <step.icon className="w-5 h-5" />
                    </div>
                    <div className="flex-1">
                      <div className="text-xs text-purple-400 mb-1">STEP {index + 1}</div>
                      <h4 className="text-lg mb-2">{step.title}</h4>
                      <p className="text-sm text-gray-400">{step.description}</p>
                    </div>
                  </div>
                </div>

                {/* Arrow connector */}
                {index < wizardSteps.length - 1 && (
                  <div className="hidden md:block absolute top-1/2 -right-3 transform -translate-y-1/2 z-10">
                    <ArrowRight className="w-6 h-6 text-purple-500/50" />
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* Benefits */}
          <div className="mt-12 pt-8 border-t border-white/10">
            <div className="grid md:grid-cols-3 gap-6 text-center">
              <div>
                <div className="text-3xl mb-2">🎯</div>
                <div className="text-sm text-gray-400">No manual YAML editing required</div>
              </div>
              <div>
                <div className="text-3xl mb-2">✨</div>
                <div className="text-sm text-gray-400">Interactive prompts with validation</div>
              </div>
              <div>
                <div className="text-3xl mb-2">⚡</div>
                <div className="text-sm text-gray-400">Get started in under a minute</div>
              </div>
            </div>
          </div>
        </div>

        {/* Example output */}
        <div className="mt-12 text-center">
          <p className="text-gray-400 mb-4">The wizard generates a complete configuration file for you:</p>
          <div className="inline-flex items-center gap-2 px-6 py-3 bg-black/40 border border-white/10 rounded-lg font-mono text-sm">
            <CheckCircle2 className="w-4 h-4 text-green-400" />
            <span className="text-gray-400">Configuration saved to</span>
            <span className="text-purple-400">nexus.yaml</span>
          </div>
        </div>
      </div>
    </section>
  );
}
