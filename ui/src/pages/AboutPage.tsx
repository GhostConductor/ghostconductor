import ReactMarkdown from 'react-markdown'
import about from '../content/about.md?raw'

export default function AboutPage() {
  return (
    <div>
      {/* Hero */}
      <div className="relative w-full">
        <img
          src="/images/gc_hero.png"
          alt="Ghost Conductor"
          className="w-full max-w-5xl mx-auto"
        />
      </div>

      {/* Content */}
      <div className="max-w-3xl mx-auto p-8">
        <article className="prose prose-invert max-w-none prose-headings:text-white prose-headings:font-semibold prose-headings:mt-8 prose-headings:mb-3 prose-p:text-gray-300 prose-li:text-gray-300 prose-strong:text-white">
          <ReactMarkdown>{about}</ReactMarkdown>
        </article>
      </div>
    </div>
  )
}
