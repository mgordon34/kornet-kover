"use client";

import React from "react";
import { format, addDays, subDays } from "date-fns";
import { CalendarIcon, ChevronLeftIcon, ChevronRightIcon } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

export function DatePicker({ date, onDateChange }: { date: Date; onDateChange: (newDate: Date) => void }) {
  const [popoverOpen, setPopoverOpen] = React.useState(false);

  const goToPreviousDay = () => {
    onDateChange(subDays(date, 1));
  };

  const goToNextDay = () => {
    onDateChange(addDays(date, 1));
  };

  const handleDateSelect = (newDate: Date | undefined) => {
    if (newDate) {
      onDateChange(newDate);
      setPopoverOpen(false); // Close the popover when a date is selected
    }
  };

  return (
    <div className="flex items-center space-x-2">
      <Button className="bg-secondary" variant="outline" size="icon" onClick={goToPreviousDay}>
        <ChevronLeftIcon className="h-4 w-4" />
      </Button>
      <Popover open={popoverOpen} onOpenChange={setPopoverOpen}>
        <PopoverTrigger asChild>
          <Button
            variant={"outline"}
            className={cn(
              "bg-secondary justify-start text-left font-normal",
              !date && "text-muted-foreground"
            )}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {format(date, "MM/dd/yy")}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0 bg-dark shadow-lg rounded-lg" align="start">
          <Calendar
            mode="single"
            selected={date}
            onSelect={handleDateSelect}
            initialFocus
            formatters={{ formatDate: (date) => format(date, "MM/dd/yy") }}
          />
        </PopoverContent>
      </Popover>
      <Button className="bg-secondary" variant="outline" size="icon" onClick={goToNextDay}>
        <ChevronRightIcon className="h-4 w-4" />
      </Button>
    </div>
  );
}

