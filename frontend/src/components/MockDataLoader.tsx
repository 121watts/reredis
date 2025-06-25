import { useState } from 'react'

interface MockDataLoaderProps {
  onLoadData: (action: 'set', key: string, value: string, silent?: boolean) => void
  isConnected: boolean
}

interface DataSet {
  name: string
  description: string
  icon: string
  count: number
  generator: () => Array<{ key: string; value: string }>
}

export const MockDataLoader = ({
  onLoadData,
  isConnected,
}: MockDataLoaderProps) => {
  const [isLoading, setIsLoading] = useState(false)
  const [selectedDataSet, setSelectedDataSet] = useState<string | null>(null)

  const generateUsers = (count: number) => {
    const users = []
    const firstNames = [
      'Alice',
      'Bob',
      'Charlie',
      'Diana',
      'Eve',
      'Frank',
      'Grace',
      'Henry',
      'Ivy',
      'Jack',
    ]
    const lastNames = [
      'Smith',
      'Johnson',
      'Williams',
      'Brown',
      'Jones',
      'Garcia',
      'Miller',
      'Davis',
      'Rodriguez',
      'Martinez',
    ]

    for (let i = 1; i <= count; i++) {
      const firstName = firstNames[Math.floor(Math.random() * firstNames.length)]
      const lastName = lastNames[Math.floor(Math.random() * lastNames.length)]
      const email = `${firstName.toLowerCase()}.${lastName.toLowerCase()}@example.com`
      const age = Math.floor(Math.random() * 50) + 18

      users.push({
        key: `user:${1000 + i}`,
        value: JSON.stringify({
          id: 1000 + i,
          name: `${firstName} ${lastName}`,
          email,
          age,
          created: new Date().toISOString(),
        }),
      })
    }
    return users
  }

  const generateSessions = (count: number) => {
    const sessions = []
    for (let i = 1; i <= count; i++) {
      const sessionId = Math.random().toString(36).substr(2, 16)
      const userId = Math.floor(Math.random() * 100) + 1000
      const expiresIn = Math.floor(Math.random() * 3600) + 1800 // 30min to 90min

      sessions.push({
        key: `session:${sessionId}`,
        value: JSON.stringify({
          userId,
          ip: `192.168.1.${Math.floor(Math.random() * 255)}`,
          userAgent: 'Mozilla/5.0 (compatible)',
          createdAt: new Date().toISOString(),
          expiresIn,
        }),
      })
    }
    return sessions
  }

  const generateCounters = (count: number) => {
    const counters = []
    const counterTypes = [
      'pageviews:homepage',
      'pageviews:about',
      'pageviews:contact',
      'api:requests',
      'api:errors',
      'downloads:app',
      'signups:daily',
      'logins:failed',
      'cache:hits',
      'cache:misses',
    ]

    for (let i = 0; i < Math.min(count, counterTypes.length); i++) {
      counters.push({
        key: counterTypes[i],
        value: Math.floor(Math.random() * 10000).toString(),
      })
    }
    return counters
  }

  const generateProducts = (count: number) => {
    const products = []
    const categories = ['electronics', 'clothing', 'books', 'home', 'sports']
    const adjectives = ['premium', 'classic', 'modern', 'vintage', 'luxury']
    const nouns = ['widget', 'gadget', 'item', 'product', 'accessory']

    for (let i = 1; i <= count; i++) {
      const category = categories[Math.floor(Math.random() * categories.length)]
      const adjective =
        adjectives[Math.floor(Math.random() * adjectives.length)]
      const noun = nouns[Math.floor(Math.random() * nouns.length)]

      products.push({
        key: `product:${i}`,
        value: JSON.stringify({
          id: i,
          name: `${adjective} ${noun}`,
          category,
          price: (Math.random() * 500 + 10).toFixed(2),
          inStock: Math.random() > 0.2,
          rating: (Math.random() * 2 + 3).toFixed(1),
        }),
      })
    }
    return products
  }

  const dataSets: DataSet[] = [
    {
      name: 'Users',
      description: 'User profiles with names, emails, and metadata',
      icon: '👥',
      count: 25,
      generator: () => generateUsers(25),
    },
    {
      name: 'Sessions',
      description: 'Active user sessions with expiration data',
      icon: '🔐',
      count: 15,
      generator: () => generateSessions(15),
    },
    {
      name: 'Counters',
      description: 'Page views, API calls, and other metrics',
      icon: '📊',
      count: 10,
      generator: () => generateCounters(10),
    },
    {
      name: 'Products',
      description: 'E-commerce product catalog with pricing',
      icon: '🛍️',
      count: 20,
      generator: () => generateProducts(20),
    },
  ]

  const loadDataSet = async (dataSet: DataSet) => {
    if (!isConnected) return

    setIsLoading(true)
    setSelectedDataSet(dataSet.name)

    try {
      const data = dataSet.generator()

      // Load data with slight delay to show progress
      for (let i = 0; i < data.length; i++) {
        const item = data[i]
        onLoadData('set', item.key, item.value, true)

        // Small delay to prevent overwhelming the UI
        if (i % 5 === 0) {
          await new Promise((resolve) => setTimeout(resolve, 50))
        }
      }
    } finally {
      setIsLoading(false)
      setSelectedDataSet(null)
    }
  }

  return (
    <div className="bg-white/5 backdrop-blur-xl p-6 rounded-3xl shadow-2xl border border-white/10">
      <div className="mb-6">
        <h2 className="text-xl font-semibold text-white mb-2">
          Load Sample Data
        </h2>
        <p className="text-slate-400 text-sm">
          Populate your cache with realistic test data for development and
          testing
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {dataSets.map((dataSet) => (
          <button
            key={dataSet.name}
            onClick={() => loadDataSet(dataSet)}
            disabled={!isConnected || isLoading}
            className="p-4 rounded-2xl bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed text-left group"
          >
            <div className="flex items-start gap-3">
              <span className="text-2xl">{dataSet.icon}</span>
              <div className="flex-1">
                <div className="flex items-center justify-between mb-1">
                  <h3 className="font-medium text-white">{dataSet.name}</h3>
                  <span className="text-xs text-slate-400 bg-slate-700/50 px-2 py-1 rounded-full">
                    {dataSet.count} items
                  </span>
                </div>
                <p className="text-sm text-slate-400">{dataSet.description}</p>
                {isLoading && selectedDataSet === dataSet.name && (
                  <div className="mt-2 flex items-center gap-2 text-xs text-purple-300">
                    <div className="w-3 h-3 border border-purple-300 border-t-transparent rounded-full animate-spin"></div>
                    Loading data...
                  </div>
                )}
              </div>
            </div>
          </button>
        ))}
      </div>

      {!isConnected && (
        <div className="mt-4 p-3 rounded-lg bg-yellow-500/20 border border-yellow-500/30 text-yellow-300 text-sm">
          ⚠️ Connect to WebSocket to load sample data
        </div>
      )}
    </div>
  )
}