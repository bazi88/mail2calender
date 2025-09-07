/**
 * Details Crawler - Extract information from schema tabs
 * This crawler can navigate to web pages and extract data from schema tabs
 */

const puppeteer = require('puppeteer');
const fs = require('fs').promises;
const path = require('path');

class DetailsCrawler {
    constructor(options = {}) {
        this.options = {
            headless: options.headless !== false, // Default to headless
            timeout: options.timeout || 30000,
            userAgent: options.userAgent || 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
            outputDir: options.outputDir || './crawler_output',
            ...options
        };
        this.browser = null;
        this.page = null;
    }

    /**
     * Initialize the browser and page
     */
    async init() {
        this.browser = await puppeteer.launch({
            headless: this.options.headless,
            args: ['--no-sandbox', '--disable-setuid-sandbox']
        });
        this.page = await this.browser.newPage();
        await this.page.setUserAgent(this.options.userAgent);
        await this.page.setViewport({ width: 1920, height: 1080 });
        
        // Create output directory if it doesn't exist
        try {
            await fs.mkdir(this.options.outputDir, { recursive: true });
        } catch (error) {
            console.warn(`Could not create output directory: ${error.message}`);
        }
    }

    /**
     * Navigate to a URL and wait for it to load
     * @param {string} url - The URL to navigate to
     */
    async navigateTo(url) {
        if (!this.page) {
            throw new Error('Crawler not initialized. Call init() first.');
        }
        
        console.log(`Navigating to: ${url}`);
        await this.page.goto(url, { 
            waitUntil: 'networkidle2',
            timeout: this.options.timeout 
        });
        
        // Wait a bit for any dynamic content to load
        await this.page.waitForTimeout(2000);
    }

    /**
     * Find and click on a schema tab (lược đồ tab)
     * @param {string} tabSelector - CSS selector for the schema tab
     */
    async clickSchemaTab(tabSelector = null) {
        if (!this.page) {
            throw new Error('Crawler not initialized. Call init() first.');
        }

        // Common selectors for schema tabs
        const commonSelectors = [
            tabSelector,
            'a[href*="schema"]',
            'button:contains("Schema")',
            'button:contains("Lược đồ")',
            '.tab:contains("Schema")',
            '.tab:contains("Lược đồ")',
            '[data-tab="schema"]',
            '[aria-label*="schema"]',
            '[title*="schema"]'
        ].filter(Boolean);

        let tabFound = false;
        
        for (const selector of commonSelectors) {
            try {
                // Wait for the element to be available
                await this.page.waitForSelector(selector, { timeout: 5000 });
                
                // Click the tab
                await this.page.click(selector);
                console.log(`Clicked schema tab using selector: ${selector}`);
                
                // Wait for tab content to load
                await this.page.waitForTimeout(2000);
                tabFound = true;
                break;
            } catch (error) {
                console.log(`Tab selector "${selector}" not found, trying next...`);
            }
        }

        if (!tabFound) {
            throw new Error('Could not find schema tab. Please provide a specific selector.');
        }
    }

    /**
     * Extract schema information from the current page
     * @param {Object} extractors - Custom extraction functions
     */
    async extractSchemaInfo(extractors = {}) {
        if (!this.page) {
            throw new Error('Crawler not initialized. Call init() first.');
        }

        const defaultExtractors = {
            // Extract text content from common schema elements
            tables: async () => {
                return await this.page.evaluate(() => {
                    const tables = document.querySelectorAll('table');
                    return Array.from(tables).map(table => {
                        const headers = Array.from(table.querySelectorAll('thead th, tr:first-child th, tr:first-child td')).map(th => th.textContent.trim());
                        const rows = Array.from(table.querySelectorAll('tbody tr, tr:not(:first-child)')).map(row => {
                            return Array.from(row.querySelectorAll('td, th')).map(cell => cell.textContent.trim());
                        });
                        return { headers, rows };
                    });
                });
            },

            // Extract structured data (JSON-LD, microdata, etc.)
            structuredData: async () => {
                return await this.page.evaluate(() => {
                    const scripts = document.querySelectorAll('script[type="application/ld+json"]');
                    return Array.from(scripts).map(script => {
                        try {
                            return JSON.parse(script.textContent);
                        } catch (e) {
                            return null;
                        }
                    }).filter(Boolean);
                });
            },

            // Extract schema-related text content
            schemaText: async () => {
                return await this.page.evaluate(() => {
                    const schemaElements = document.querySelectorAll(
                        '.schema, .schema-info, .schema-content, .schema-details, ' +
                        '[class*="schema"], [id*="schema"], [data-schema]'
                    );
                    return Array.from(schemaElements).map(el => ({
                        tag: el.tagName.toLowerCase(),
                        className: el.className,
                        id: el.id,
                        text: el.textContent.trim(),
                        html: el.innerHTML
                    }));
                });
            },

            // Extract all visible text from the page
            allText: async () => {
                return await this.page.evaluate(() => {
                    return document.body.textContent.trim();
                });
            },

            // Extract meta information
            metadata: async () => {
                return await this.page.evaluate(() => {
                    const title = document.title;
                    const description = document.querySelector('meta[name="description"]')?.content || '';
                    const url = window.location.href;
                    const timestamp = new Date().toISOString();
                    return { title, description, url, timestamp };
                });
            }
        };

        // Merge custom extractors with default ones
        const allExtractors = { ...defaultExtractors, ...extractors };
        const extractedData = {};

        console.log('Extracting schema information...');
        
        for (const [key, extractor] of Object.entries(allExtractors)) {
            try {
                console.log(`Extracting: ${key}`);
                extractedData[key] = await extractor();
            } catch (error) {
                console.warn(`Failed to extract ${key}: ${error.message}`);
                extractedData[key] = null;
            }
        }

        return extractedData;
    }

    /**
     * Save extracted data to a file
     * @param {Object} data - The data to save
     * @param {string} filename - The filename (without extension)
     */
    async saveData(data, filename = 'schema_data') {
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        const outputPath = path.join(this.options.outputDir, `${filename}_${timestamp}.json`);
        
        try {
            await fs.writeFile(outputPath, JSON.stringify(data, null, 2), 'utf8');
            console.log(`Data saved to: ${outputPath}`);
            return outputPath;
        } catch (error) {
            console.error(`Failed to save data: ${error.message}`);
            throw error;
        }
    }

    /**
     * Take a screenshot of the current page
     * @param {string} filename - The filename for the screenshot
     */
    async takeScreenshot(filename = 'schema_page') {
        if (!this.page) {
            throw new Error('Crawler not initialized. Call init() first.');
        }

        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        const screenshotPath = path.join(this.options.outputDir, `${filename}_${timestamp}.png`);
        
        try {
            await this.page.screenshot({ 
                path: screenshotPath, 
                fullPage: true 
            });
            console.log(`Screenshot saved to: ${screenshotPath}`);
            return screenshotPath;
        } catch (error) {
            console.error(`Failed to take screenshot: ${error.message}`);
            throw error;
        }
    }

    /**
     * Close the browser
     */
    async close() {
        if (this.browser) {
            await this.browser.close();
            this.browser = null;
            this.page = null;
        }
    }

    /**
     * Main crawling method - combines all steps
     * @param {string} url - The URL to crawl
     * @param {string} tabSelector - CSS selector for the schema tab (optional)
     * @param {Object} extractors - Custom extraction functions (optional)
     */
    async crawl(url, tabSelector = null, extractors = {}) {
        try {
            await this.init();
            await this.navigateTo(url);
            
            // Try to click schema tab if it exists
            try {
                await this.clickSchemaTab(tabSelector);
            } catch (error) {
                console.warn(`Could not find/click schema tab: ${error.message}`);
                console.log('Proceeding with current page content...');
            }
            
            // Extract data
            const data = await this.extractSchemaInfo(extractors);
            
            // Take screenshot
            await this.takeScreenshot();
            
            // Save data
            await this.saveData(data);
            
            return data;
        } catch (error) {
            console.error(`Crawling failed: ${error.message}`);
            throw error;
        } finally {
            await this.close();
        }
    }
}

// Example usage function
async function exampleUsage() {
    const crawler = new DetailsCrawler({
        headless: false, // Set to true for production
        outputDir: './crawler_results'
    });

    try {
        // Example: crawl a website with schema information
        const data = await crawler.crawl('https://example.com', null, {
            // Custom extractor for specific information
            customData: async () => {
                return await crawler.page.evaluate(() => {
                    // Add custom extraction logic here
                    return { message: 'Add your custom extraction logic here' };
                });
            }
        });

        console.log('Extracted data:', data);
    } catch (error) {
        console.error('Error:', error);
    }
}

// Export for use as a module
module.exports = DetailsCrawler;

// If run directly, execute example
if (require.main === module) {
    console.log('Details Crawler - Schema Information Extractor');
    console.log('Usage: node details_crawler.js');
    console.log('Modify the exampleUsage() function to specify your target URL and extraction requirements.');
    
    // Uncomment the next line to run the example
    // exampleUsage();
}