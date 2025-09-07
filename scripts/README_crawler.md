# Details Crawler - Schema Information Extractor

A powerful web crawler designed to extract information from schema tabs (lược đồ) on web pages. This crawler uses Puppeteer to automate browser interactions and extract structured data.

## Features

- **Tab Navigation**: Automatically finds and clicks on schema tabs using various selectors
- **Flexible Extraction**: Extract tables, structured data, metadata, and custom content
- **Vietnamese Support**: Supports Vietnamese websites and schema tabs ("lược đồ")
- **Screenshot Capture**: Takes screenshots of the crawled pages
- **Data Export**: Saves extracted data in JSON format with timestamps
- **Customizable**: Easy to extend with custom extraction functions

## Installation

1. Navigate to the scripts directory:
```bash
cd scripts
```

2. Install dependencies:
```bash
npm install
```

## Usage

### Basic Usage

```javascript
const DetailsCrawler = require('./details_crawler');

const crawler = new DetailsCrawler({
    headless: true,        // Set to false to see browser
    outputDir: './results' // Where to save output files
});

// Crawl a website and extract schema information
const data = await crawler.crawl('https://example.com');
```

### Advanced Usage with Custom Extractors

```javascript
const data = await crawler.crawl('https://example.com', 
    // Tab selector (optional) - will try common selectors if not provided
    'button:contains("Schema")',
    
    // Custom extractors
    {
        productInfo: async () => {
            return await crawler.page.evaluate(() => {
                // Your custom extraction logic here
                const products = document.querySelectorAll('.product');
                return Array.from(products).map(p => ({
                    name: p.querySelector('.name')?.textContent,
                    price: p.querySelector('.price')?.textContent
                }));
            });
        }
    }
);
```

### Vietnamese Website Example

```javascript
const crawler = new DetailsCrawler();

const data = await crawler.crawl('https://example.vn',
    // Try Vietnamese schema tab selectors
    'a:contains("Lược đồ"), button:contains("Lược đồ")',
    
    // Custom Vietnamese content extractor
    {
        vietnameseData: async () => {
            return await crawler.page.evaluate(() => {
                // Extract Vietnamese schema information
                const elements = document.querySelectorAll('[class*="schema"], [data-schema]');
                return Array.from(elements).map(el => el.textContent);
            });
        }
    }
);
```

## Configuration Options

```javascript
const crawler = new DetailsCrawler({
    headless: true,                    // Run in headless mode
    timeout: 30000,                    // Page load timeout (ms)
    outputDir: './crawler_output',     // Output directory
    userAgent: 'Custom User Agent'     // Custom user agent
});
```

## Output Files

The crawler creates several output files:

- **JSON Data**: `schema_data_TIMESTAMP.json` - Extracted information
- **Screenshots**: `schema_page_TIMESTAMP.png` - Page screenshots

## Default Extractors

The crawler includes several built-in extractors:

1. **tables**: Extracts all table data (headers and rows)
2. **structuredData**: Finds JSON-LD structured data
3. **schemaText**: Extracts schema-related text content
4. **allText**: Gets all visible text from the page
5. **metadata**: Extracts page title, description, URL, and timestamp

## Common Schema Tab Selectors

The crawler automatically tries these selectors to find schema tabs:

- `a[href*="schema"]`
- `button:contains("Schema")`
- `button:contains("Lược đồ")`
- `.tab:contains("Schema")`
- `[data-tab="schema"]`
- `[aria-label*="schema"]`

## Running Tests

Run the test suite to verify the crawler works:

```bash
npm test
```

Or run the test file directly:

```bash
node test_crawler.js
```

## Examples

### Extract E-commerce Product Schema

```javascript
const crawler = new DetailsCrawler();

const productData = await crawler.crawl('https://shop.example.com/product/123', null, {
    product: async () => {
        return await crawler.page.evaluate(() => {
            return {
                name: document.querySelector('[itemprop="name"]')?.textContent,
                price: document.querySelector('[itemprop="price"]')?.textContent,
                description: document.querySelector('[itemprop="description"]')?.textContent,
                image: document.querySelector('[itemprop="image"]')?.src
            };
        });
    }
});
```

### Extract Event Schema

```javascript
const eventData = await crawler.crawl('https://events.example.com/event/456', null, {
    event: async () => {
        return await crawler.page.evaluate(() => {
            return {
                name: document.querySelector('[itemprop="name"]')?.textContent,
                startDate: document.querySelector('[itemprop="startDate"]')?.getAttribute('datetime'),
                location: document.querySelector('[itemprop="location"]')?.textContent,
                organizer: document.querySelector('[itemprop="organizer"]')?.textContent
            };
        });
    }
});
```

## Troubleshooting

### Common Issues

1. **Schema tab not found**: Provide a specific selector for your target website
2. **Timeout errors**: Increase the timeout value in options
3. **Permission errors**: Ensure the output directory is writable

### Debug Mode

Run with headless: false to see what the crawler is doing:

```javascript
const crawler = new DetailsCrawler({ headless: false });
```

## Error Handling

The crawler includes comprehensive error handling:

```javascript
try {
    const data = await crawler.crawl('https://example.com');
    console.log('Success:', data);
} catch (error) {
    console.error('Crawling failed:', error.message);
}
```

## Contributing

To add new features or improve the crawler:

1. Modify `details_crawler.js` for core functionality
2. Add tests in `test_crawler.js`
3. Update this README with new features

## License

MIT License - see the project root for license details.