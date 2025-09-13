import { HeroSection } from '../components/hero-section'
import { FeaturesSection } from '../components/features-section'
import { Footer } from '../components/footer'

function HomePage() {
  return (
    <main className="min-h-screen bg-gradient-to-br from-gray-900 via-indigo-900 to-blue-900 text-white flex flex-col">
      <HeroSection />
      <FeaturesSection />
      <Footer />
    </main>
  )
}

export { HomePage as default }
