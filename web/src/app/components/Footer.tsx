import { Terminal, Github, Twitter, Heart } from 'lucide-react';

export function Footer() {
	const currentYear = new Date().getFullYear();

	const handleSmoothScroll = (
		e: React.MouseEvent<HTMLAnchorElement>,
		targetId: string,
	) => {
		e.preventDefault();
		const element = document.getElementById(targetId);
		if (element) {
			element.scrollIntoView({ behavior: 'smooth', block: 'start' });
		}
	};

	return (
		<footer className="border-t border-white/10 py-12 px-6">
			<div className="max-w-7xl mx-auto">
				<div className="grid grid-cols-1 md:grid-cols-4 gap-12 mb-12">
					{/* Brand */}
					<div className="md:col-span-2">
						<div className="flex items-center gap-2 mb-4">
							<div className="w-8 h-8 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center">
								<Terminal className="w-5 h-5" />
							</div>
							<span className="text-xl">Nexus</span>
						</div>
						<p className="text-gray-400 leading-relaxed max-w-md">
							Turn scattered commands into a clean, interactive interface. The
							developer tool that brings simplicity to your terminal workflow.
						</p>
					</div>

					{/* Product */}
					<div>
						<h4 className="mb-4 text-sm uppercase tracking-wider text-gray-500">
							Product
						</h4>
						<ul className="space-y-3">
							<li>
								<a
									href="#features"
									className="text-gray-400 hover:text-white transition-colors"
									onClick={(e) => handleSmoothScroll(e, 'features')}
								>
									Features
								</a>
							</li>
							<li>
								<a
									href="#how-it-works"
									className="text-gray-400 hover:text-white transition-colors"
									onClick={(e) => handleSmoothScroll(e, 'how-it-works')}
								>
									How it Works
								</a>
							</li>
							<li>
								<a
									href="#wizard"
									className="text-gray-400 hover:text-white transition-colors"
									onClick={(e) => handleSmoothScroll(e, 'wizard')}
								>
									Wizard
								</a>
							</li>
							<li>
								<a
									href="#documentation"
									className="text-gray-400 hover:text-white transition-colors"
									onClick={(e) => handleSmoothScroll(e, 'documentation')}
								>
									Documentation
								</a>
							</li>
							<li>
								<a
									href="#examples"
									className="text-gray-400 hover:text-white transition-colors"
									onClick={(e) => handleSmoothScroll(e, 'examples')}
								>
									Examples
								</a>
							</li>
						</ul>
					</div>

					{/* Resources */}
					<div>
						<h4 className="mb-4 text-sm uppercase tracking-wider text-gray-500">
							Resources
						</h4>
						<ul className="space-y-3">
							<li>
								<a
									href="https://github.com/sandro-sikic/nexus"
									className="text-gray-400 hover:text-white transition-colors"
								>
									GitHub
								</a>
							</li>
							<li>
								<a
									href="https://github.com/sandro-sikic/nexus/releases"
									className="text-gray-400 hover:text-white transition-colors"
								>
									Releases
								</a>
							</li>
							<li>
								<a
									href="https://github.com/sandro-sikic/nexus/blob/main/LICENSE"
									className="text-gray-400 hover:text-white transition-colors"
								>
									License
								</a>
							</li>
						</ul>
					</div>
				</div>

				{/* Bottom bar */}
				<div className="pt-8 border-t border-white/10 flex flex-col md:flex-row justify-between items-center gap-4">
					<div className="flex items-center gap-2 text-sm text-gray-400">
						<span>© {currentYear} Nexus.</span>
						<span>Made with</span>
						<Heart className="w-4 h-4 text-red-400 fill-red-400" />
						<span>by the open-source community</span>
					</div>

					<div className="flex items-center gap-4">
						<a
							href="https://github.com/sandro-sikic"
							className="w-10 h-10 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 flex items-center justify-center transition-all hover:border-white/20"
							aria-label="GitHub"
						>
							<Github className="w-5 h-5" />
						</a>
					</div>
				</div>
			</div>
		</footer>
	);
}
