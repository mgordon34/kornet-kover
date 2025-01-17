"use client";

import React, { useState, useEffect } from "react";
import { StrategyPicks } from "../types";
import PickList from "../components/PickList";
import { DatePicker } from "../components/DatePicker";
import { calculateDiff } from "../../lib/pick_utils";

interface ClientPicksProps {
  initialDate: Date;
  initialStrategies: StrategyPicks[];
}

const ClientPicks: React.FC<ClientPicksProps> = ({ initialDate, initialStrategies }) => {
  const [selectedDate, setSelectedDate] = useState(initialDate);
  const [strategies, setStrategies] = useState<StrategyPicks[]>(initialStrategies);

  const fetchPicks = async (date: Date) => {
    const userId = 1;
    const formattedDate = new Intl.DateTimeFormat("en-CA").format(date);
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/prop-picks?user_id=${userId}&date=${formattedDate}`, {
        method: "GET",
        cache: "no-store",
      });
      const data: StrategyPicks[] = await res.json();
      setStrategies(data);
    } catch (error) {
      console.error("Error fetching picks:", error);
    }
  };

  useEffect(() => {
    if (selectedDate) {
      fetchPicks(selectedDate);
    }
  }, [selectedDate]);

  return (
    <div className="flex flex-col items-center justify-items-center p-8">
      <div className="absolute right-8 top-24">
        <DatePicker date={selectedDate} onDateChange={setSelectedDate} />
      </div>
      <div className="inline-flex flex-col">
        {strategies && strategies.length > 0 ? (
          strategies.map((strategy) => {
            const sortedPicks = strategy.picks.sort((a, b) => calculateDiff(b) - calculateDiff(a));
            return (
              <div key={strategy.strat_id} className="my-4">
                <h1 className="text-2xl font-semibold">{strategy.strat_name}</h1>
                <PickList picks={sortedPicks} />
              </div>
            );
          })
        ) : (
          <div className="text-center">
            <p>No picks available for the selected date.</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default ClientPicks;
