import { MetricsCard } from '@/components/MetricsCard';

export default function Home() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 to-slate-800">
      <div className="text-center px-4">
        <h1 className="text-6xl font-bold text-white mb-4">
          🌙 LunaSentri
        </h1>
        <p className="text-xl text-slate-300 mb-8">
          Lightweight Server Monitoring Dashboard
        </p>
        <MetricsCard />
      </div>
    </div>
  );
}
