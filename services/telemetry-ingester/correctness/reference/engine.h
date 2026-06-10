#pragma once

#include <string>

// process_order_line takes a single JSON order line from stdin
// and returns a JSON fill string for stdout.
// Day 5: will implement full price-time priority FIFO matching.
// Today: returns a trivially-correct fill (price=0, quantity=0, side="").
std::string process_order_line(const std::string& line);
