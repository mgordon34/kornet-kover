"use client"

import * as React from "react"
import { format, addDays, subDays } from "date-fns"
import { CalendarIcon, ChevronLeftIcon, ChevronRightIcon } from 'lucide-react'

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"

export function DatePicker() {
  const [date, setDate] = React.useState<Date>(new Date())
  const [popoverOpen, setPopoverOpen] = React.useState(false)

  const goToPreviousDay = () => {
    setDate((prevDate) => subDays(prevDate, 1))
  }

  const goToNextDay = () => {
    setDate((prevDate) => addDays(prevDate, 1))
  }

  const handleDateSelect = (newDate: Date | null) => {
    if (newDate) {
      setDate(newDate)
      setPopoverOpen(false) // Close the popover when a date is selected
    }
  }

  return (
    <div className="flex items-center space-x-2">
      <Button variant="outline" size="icon" onClick={goToPreviousDay}>
        <ChevronLeftIcon className="h-4 w-4" />
      </Button>
      <Popover open={popoverOpen} onOpenChange={setPopoverOpen}>
        <PopoverTrigger asChild>
          <Button
            variant={"outline"}
            className={cn(
              "justify-start text-left font-normal",
              !date && "text-muted-foreground"
            )}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {format(date, "MM/dd/yy")}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            selected={date}
            onSelect={handleDateSelect}
            initialFocus
            formatters={{ formatDate: (date) => format(date, "MM/dd/yy") }}
          />
        </PopoverContent>
      </Popover>
      <Button variant="outline" size="icon" onClick={goToNextDay}>
        <ChevronRightIcon className="h-4 w-4" />
      </Button>
    </div>
  )
}
