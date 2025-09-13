function HeroSection() {
  return (
    <section className="flex flex-col items-center justify-center py-24 px-4 text-center">
      <h1 className="text-5xl md:text-7xl font-extrabold bg-gradient-to-r from-indigo-400 via-blue-400 to-cyan-400 bg-clip-text text-transparent mb-6 drop-shadow-lg">
        Cotai: IA Copilot para Equipes
      </h1>
      <p className="max-w-2xl text-lg md:text-2xl text-gray-200 mb-8">
        Plataforma de IA colaborativa para times modernos. Integre, converse, crie e automatize fluxos com segurança e privacidade.
      </p>
      <a href="#features" className="inline-block px-8 py-3 rounded-full bg-indigo-600 hover:bg-indigo-700 text-white font-semibold shadow-lg transition-all">
        Conheça os recursos
      </a>
    </section>
  )
}

export { HeroSection }
