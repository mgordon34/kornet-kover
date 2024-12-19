// app/api/strategies/route.ts
import { NextResponse } from 'next/server';

// Define the Strategy interface
interface Strategy {
  id: number;
  name: string;
}

export async function GET() {
  // Mocking example strategy data (replace with real database query later)
  const strategies: Strategy[] = [
    { id: 1, name: 'Risk Management Strategy' },
    { id: 2, name: 'Momentum Strategy' },
    { id: 3, name: 'Value Betting Strategy' }
  ];

  // Return strategies as JSON
  return NextResponse.json(strategies);
}
