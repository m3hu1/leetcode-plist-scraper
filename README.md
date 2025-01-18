# LeetCode Problem List Scraper

A simple command-line tool to scrape LeetCode problem lists and save them as JSON files.

## Usage

```bash
go run scraper.go <list-id> <name>

# Example
go run scraper.go ajpcecv6 "Blind 75"
```

This will create a JSON file in the `data` directory containing the problem list details.

## Output Format
```json
{
  "name": "List Name",
  "problems": [
    {
      "id": "problem-slug",
      "name": "Problem Title",
      "link": "https://leetcode.com/problems/problem-slug",
      "difficulty": "easy|medium|hard"
    }
  ]
}
```
