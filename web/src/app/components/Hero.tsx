import { Terminal, Download, Github } from 'lucide-react';
import { motion } from 'motion/react';
import { useMemo } from 'react';

const backgroundCommands = [
  'npm run dev',
  'docker compose up',
  'git push origin main',
  'npm install',
  'yarn build',
  'kubectl get pods',
  'docker logs -f',
  'npm run test',
  'cargo run',
  'go build',
  'python manage.py runserver',
  'rails server',
  'terraform apply',
  'make build',
  'gradle build',
  'pnpm install',
  'npm run deploy',
  'git pull --rebase',
  'docker-compose down',
  'npm run lint',
  'yarn test:watch',
  'docker build -t app .',
  'kubectl apply -f',
  'npm run start',
  'bundle exec rails s',
  'cargo test',
  'go test ./...',
  'pytest',
  'mvn clean install',
  'npm run typecheck',
  'docker exec -it',
  'git checkout -b',
  'heroku logs --tail',
  'yarn dev',
  'git commit -m',
  'npm run build',
  'docker ps -a',
  'kubectl describe pod',
  'mix phx.server',
  'dotnet run',
  'pip install -r requirements.txt',
  'composer install',
  'ansible-playbook deploy.yml',
  'rsync -avz',
  'ssh user@server',
  'scp file.txt',
  'git status',
  'npm audit fix',
  'yarn upgrade',
  'docker rm -f',
  'kubectl logs -f',
  'go mod tidy',
  'cargo build --release',
  'bundle install',
  'rails db:migrate',
  'django-admin startproject',
  'php artisan serve',
  'ng serve',
  'flutter run',
  'swiftpm build',
  'gradle test',
  'mvn package',
  'npm ci',
  'pnpm dev',
];

// Check if two rectangles overlap
function checkOverlap(x1: number, y1: number, w1: number, h1: number, 
                      x2: number, y2: number, w2: number, h2: number, 
                      padding: number = 5) {
  return !(x1 + w1 + padding < x2 || 
           x2 + w2 + padding < x1 || 
           y1 + h1 + padding < y2 || 
           y2 + h2 + padding < y1);
}

export function Hero() {
  // Generate non-overlapping positions
  const commandPositions = useMemo(() => {
    const positions: Array<{
      x: number;
      y: number;
      rotate: number;
      delay: number;
      duration: number;
      width: number;
      height: number;
    }> = [];

    backgroundCommands.forEach((cmd) => {
      let attempts = 0;
      let positionFound = false;
      
      // Estimate command width based on character count (rough approximation)
      const estimatedWidth = cmd.length * 0.6; // percentage units
      const estimatedHeight = 3; // percentage units
      
      while (!positionFound && attempts < 100) {
        const randomX = Math.random() * (100 - estimatedWidth);
        const randomY = Math.random() * (100 - estimatedHeight);
        const randomRotate = Math.random() * 20 - 10;
        const randomDelay = Math.random() * 2;
        const randomDuration = 20 + Math.random() * 10;
        
        // Check if this position overlaps with any existing position
        const overlaps = positions.some(pos => 
          checkOverlap(
            randomX, randomY, estimatedWidth, estimatedHeight,
            pos.x, pos.y, pos.width, pos.height
          )
        );
        
        if (!overlaps) {
          positions.push({
            x: randomX,
            y: randomY,
            rotate: randomRotate,
            delay: randomDelay,
            duration: randomDuration,
            width: estimatedWidth,
            height: estimatedHeight,
          });
          positionFound = true;
        }
        
        attempts++;
      }
    });
    
    return positions;
  }, []);

  return (
    <section className="relative min-h-screen flex items-center justify-center px-6 py-20 overflow-hidden">
      {/* Gradient background effect */}
      <div className="absolute inset-0 bg-gradient-to-b from-purple-900/20 via-transparent to-transparent pointer-events-none" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-purple-900/20 via-transparent to-transparent pointer-events-none" />
      
      {/* Scattered background commands */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        {backgroundCommands.map((cmd, index) => {
          const pos = commandPositions[index];
          if (!pos) return null;
          
          return (
            <motion.div
              key={index}
              initial={{ opacity: 0, y: 20 }}
              animate={{ 
                opacity: [0, 0.25, 0.25, 0],
                y: [pos.y + '%', (pos.y - 20) + '%']
              }}
              transition={{
                duration: pos.duration,
                delay: pos.delay,
                repeat: Infinity,
                repeatType: 'loop',
              }}
              className="absolute font-mono text-sm md:text-base text-gray-300 whitespace-nowrap"
              style={{
                left: `${pos.x}%`,
                top: `${pos.y}%`,
                transform: `rotate(${pos.rotate}deg)`,
              }}
            >
              $ {cmd}
            </motion.div>
          );
        })}
      </div>
      
      <div className="max-w-6xl mx-auto text-center relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="mb-8"
        >
          <div className="inline-flex items-center gap-3 px-4 py-2 rounded-full bg-purple-500/10 border border-purple-500/20 mb-8">
            <Terminal className="w-4 h-4 text-purple-400" />
            <span className="text-sm text-purple-300">Open Source Developer Tool</span>
          </div>
        </motion.div>

        <motion.h1
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.1 }}
          className="text-6xl md:text-7xl lg:text-8xl mb-6 tracking-tight"
        >
          Your commands.
          <br />
          <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 via-pink-400 to-purple-400">
            One interface.
          </span>
        </motion.h1>

        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="text-xl md:text-2xl text-gray-400 mb-12 max-w-3xl mx-auto leading-relaxed"
        >
          Stop memorizing commands. Nexus transforms your scattered terminal commands into a clean, interactive TUI that works anywhere.
        </motion.p>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.3 }}
          className="flex flex-col sm:flex-row gap-4 justify-center items-center"
        >
          <button className="group relative inline-flex items-center gap-2 px-8 py-4 bg-gradient-to-r from-purple-600 to-pink-600 rounded-lg overflow-hidden transition-all hover:scale-105">
            <div className="absolute inset-0 bg-gradient-to-r from-purple-500 to-pink-500 opacity-0 group-hover:opacity-100 transition-opacity" />
            <Download className="w-5 h-5 relative z-10" />
            <span className="text-lg relative z-10">Download Nexus</span>
          </button>
          
          <button className="inline-flex items-center gap-2 px-8 py-4 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg transition-all hover:border-white/20 backdrop-blur-sm">
            <Github className="w-5 h-5" />
            <span className="text-lg">View on GitHub</span>
          </button>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.4 }}
          className="mt-12 text-sm text-gray-500"
        >
          <p>Cross-platform • Windows, macOS, Linux</p>
        </motion.div>
      </div>
    </section>
  );
}