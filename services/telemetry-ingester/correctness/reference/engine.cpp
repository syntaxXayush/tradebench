#include "engine.h"

#include <iostream>
#include <string>

// process_order_line returns a trivially-correct fill JSON for any input order.
// Day 5: will parse the order JSON, run it through the price-time priority
// matching engine, and return the computed expected fill.
std::string process_order_line(const std::string& line) {
    // Stub: ignore input, return a structurally valid fill with zero values.
    return "{\"price\":0.0,\"quantity\":0.0,\"side\":\"\"}";
}

int main() {
    // Read order lines from stdin until EOF.
    // For each line, write one JSON fill object to stdout.
    std::string line;
    while (std::getline(std::cin, line)) {
        if (line.empty()) {
            continue;
        }
        std::cout << process_order_line(line) << "\n";
    }
    return 0;
}
