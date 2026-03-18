import { Hero } from './components/Hero';
import { Features } from './components/Features';
import { HowItWorks } from './components/HowItWorks';
import { Wizard } from './components/Wizard';
import { CodeExample } from './components/CodeExample';
import { TUIPreview } from './components/TUIPreview';
import { RealWorldExample } from './components/RealWorldExample';
import { CTA } from './components/CTA';
import { Footer } from './components/Footer';

export default function App() {
  return (
    <div className="min-h-screen bg-[#0a0a0a] text-white">
      {/* Background grid effect */}
      <div className="fixed inset-0 bg-[linear-gradient(to_right,#ffffff05_1px,transparent_1px),linear-gradient(to_bottom,#ffffff05_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)] pointer-events-none" />
      
      <div className="relative z-10">
        <Hero />
        <Features />
        <HowItWorks />
        <Wizard />
        <TUIPreview />
        <CodeExample />
        <RealWorldExample />
        <CTA />
        <Footer />
      </div>
    </div>
  );
}