function FeaturesSection() {
  return (
    <section id="features" className="py-20 px-4 bg-gray-950">
      <div className="max-w-5xl mx-auto grid gap-12 md:grid-cols-3">
        <div className="bg-gray-900 rounded-xl p-8 shadow-lg border border-indigo-800">
          <h2 className="text-xl font-bold mb-4 text-indigo-400">Integração Simples</h2>
          <p className="text-gray-300">Conecte suas ferramentas favoritas com poucos cliques e centralize o conhecimento do seu time.</p>
        </div>
        <div className="bg-gray-900 rounded-xl p-8 shadow-lg border border-blue-800">
          <h2 className="text-xl font-bold mb-4 text-blue-400">Colaboração em Tempo Real</h2>
          <p className="text-gray-300">Chat, automações e fluxos colaborativos para equipes remotas e híbridas.</p>
        </div>
        <div className="bg-gray-900 rounded-xl p-8 shadow-lg border border-cyan-800">
          <h2 className="text-xl font-bold mb-4 text-cyan-400">Privacidade & Segurança</h2>
          <p className="text-gray-300">Dados protegidos, controle granular de acesso e conformidade com LGPD.</p>
        </div>
      </div>
    </section>
  )
}

export { FeaturesSection }
