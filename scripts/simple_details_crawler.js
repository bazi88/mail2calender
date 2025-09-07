/**
 * Simple Details Crawler - Extract information from schema tabs without browser automation
 * This is a lighter version that uses HTTP requests and HTML parsing
 */

const axios = require('axios');
const cheerio = require('cheerio');
const fs = require('fs').promises;
const path = require('path');
const url = require('url');

class SimpleDetailsCrawler {
    constructor(options = {}) {
        this.options = {
            timeout: options.timeout || 30000,
            userAgent: options.userAgent || 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
            outputDir: options.outputDir || './crawler_output',
            maxRedirects: options.maxRedirects || 5,
            headers: options.headers || {},
            ...options
        };
        
        this.axiosConfig = {
            timeout: this.options.timeout,
            maxRedirects: this.options.maxRedirects,
            headers: {
                'User-Agent': this.options.userAgent,
                'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8',
                'Accept-Language': 'vi-VN,vi;q=0.9,en;q=0.8',
                'Accept-Encoding': 'gzip, deflate, br',
                'DNT': '1',
                'Connection': 'keep-alive',
                'Upgrade-Insecure-Requests': '1',
                ...this.options.headers
            }
        };
    }

    /**
     * Initialize the crawler (create output directory)
     */
    async init() {
        try {
            await fs.mkdir(this.options.outputDir, { recursive: true });
        } catch (error) {
            console.warn(`Could not create output directory: ${error.message}`);
        }
    }

    /**
     * Fetch HTML content from a URL
     * @param {string} targetUrl - The URL to fetch
     */
    async fetchPage(targetUrl) {
        console.log(`Fetching: ${targetUrl}`);
        
        try {
            const response = await axios.get(targetUrl, this.axiosConfig);
            return {
                html: response.data,
                status: response.status,
                headers: response.headers,
                url: response.config.url || targetUrl
            };
        } catch (error) {
            console.error(`Failed to fetch ${targetUrl}: ${error.message}`);
            throw error;
        }
    }

    /**
     * Find schema-related links in the HTML
     * @param {string} html - HTML content
     * @param {string} baseUrl - Base URL for resolving relative links
     */
    findSchemaLinks(html, baseUrl) {
        const $ = cheerio.load(html);
        const schemaLinks = [];

        // Look for links that might lead to schema pages
        const selectors = [
            'a[href*="schema"]',
            'a[href*="luoc-do"]',
            'a[href*="lược-đồ"]',
            'a:contains("Schema")',
            'a:contains("Lược đồ")',
            'a:contains("Structured Data")',
            'a[data-tab="schema"]',
            'button[data-tab="schema"]'
        ];

        selectors.forEach(selector => {
            $(selector).each((i, element) => {
                const href = $(element).attr('href');
                const text = $(element).text().trim();
                
                if (href) {
                    const fullUrl = url.resolve(baseUrl, href);
                    schemaLinks.push({
                        url: fullUrl,
                        text: text,
                        selector: selector
                    });
                }
            });
        });

        return schemaLinks;
    }

    /**
     * Extract schema information from HTML
     * @param {string} html - HTML content
     * @param {Object} customExtractors - Custom extraction functions
     */
    extractSchemaInfo(html, customExtractors = {}) {
        const $ = cheerio.load(html);
        
        const defaultExtractors = {
            // Extract JSON-LD structured data
            jsonLd: () => {
                const jsonLdData = [];
                $('script[type="application/ld+json"]').each((i, element) => {
                    try {
                        const data = JSON.parse($(element).html());
                        jsonLdData.push(data);
                    } catch (e) {
                        console.warn(`Invalid JSON-LD found: ${e.message}`);
                    }
                });
                return jsonLdData;
            },

            // Extract microdata
            microdata: () => {
                const microdataItems = [];
                $('[itemtype]').each((i, element) => {
                    const $element = $(element);
                    const item = {
                        type: $element.attr('itemtype'),
                        properties: {}
                    };

                    $element.find('[itemprop]').each((j, prop) => {
                        const $prop = $(prop);
                        const propName = $prop.attr('itemprop');
                        const propValue = $prop.attr('content') || $prop.text().trim();
                        item.properties[propName] = propValue;
                    });

                    microdataItems.push(item);
                });
                return microdataItems;
            },

            // Extract tables
            tables: () => {
                const tables = [];
                $('table').each((i, table) => {
                    const $table = $(table);
                    const headers = [];
                    const rows = [];

                    $table.find('thead th, tr:first-child th, tr:first-child td').each((j, header) => {
                        headers.push($(header).text().trim());
                    });

                    $table.find('tbody tr, tr:not(:first-child)').each((j, row) => {
                        const rowData = [];
                        $(row).find('td, th').each((k, cell) => {
                            rowData.push($(cell).text().trim());
                        });
                        if (rowData.length > 0) {
                            rows.push(rowData);
                        }
                    });

                    tables.push({ headers, rows });
                });
                return tables;
            },

            // Extract schema-related content
            schemaContent: () => {
                const schemaElements = [];
                const selectors = [
                    '.schema', '.schema-info', '.schema-content', '.schema-details',
                    '[class*="schema"]', '[id*="schema"]', '[data-schema]'
                ];

                selectors.forEach(selector => {
                    $(selector).each((i, element) => {
                        const $element = $(element);
                        schemaElements.push({
                            tag: element.tagName.toLowerCase(),
                            className: $element.attr('class') || '',
                            id: $element.attr('id') || '',
                            text: $element.text().trim(),
                            html: $element.html()
                        });
                    });
                });

                return schemaElements;
            },

            // Extract meta information
            metadata: () => {
                return {
                    title: $('title').text().trim(),
                    description: $('meta[name="description"]').attr('content') || '',
                    keywords: $('meta[name="keywords"]').attr('content') || '',
                    ogTitle: $('meta[property="og:title"]').attr('content') || '',
                    ogDescription: $('meta[property="og:description"]').attr('content') || '',
                    ogImage: $('meta[property="og:image"]').attr('content') || '',
                    timestamp: new Date().toISOString()
                };
            },

            // Extract product information (Vietnamese e-commerce)
            productInfo: () => {
                const product = {};
                
                // Common product selectors for Vietnamese sites
                product.name = $('.product-name, .product-title, h1, [data-view-id="pdp_product_name"]').first().text().trim();
                product.price = $('.product-price, .price, .price-sale, [data-view-id="pdp_price"]').first().text().trim();
                product.description = $('.product-description, .description, [data-view-id="pdp_product_description"]').first().text().trim();
                product.brand = $('.brand, .product-brand').first().text().trim();
                product.category = $('.breadcrumb, .category').first().text().trim();
                product.rating = $('.rating, .stars, [data-view-id="pdp_rating"]').first().text().trim();
                product.availability = $('.availability, .stock-status').first().text().trim();
                
                // Extract images
                product.images = [];
                $('.product-image img, .gallery img').each((i, img) => {
                    const src = $(img).attr('src');
                    if (src) product.images.push(src);
                });

                return product;
            },

            // Extract event information
            eventInfo: () => {
                const event = {};
                
                event.name = $('.event-name, .event-title, h1').first().text().trim();
                event.date = $('.event-date, .date').first().text().trim();
                event.time = $('.event-time, .time').first().text().trim();
                event.venue = $('.event-venue, .venue').first().text().trim();
                event.location = $('.event-location, .location').first().text().trim();
                event.description = $('.event-description, .description').first().text().trim();
                event.organizer = $('.event-organizer, .organizer').first().text().trim();
                event.price = $('.event-price, .ticket-price, .price').first().text().trim();

                return event;
            }
        };

        // Merge custom extractors
        const allExtractors = { ...defaultExtractors, ...customExtractors };
        const extractedData = {};

        console.log('Extracting schema information...');
        
        for (const [key, extractor] of Object.entries(allExtractors)) {
            try {
                console.log(`Extracting: ${key}`);
                extractedData[key] = extractor();
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
     * Main crawling method
     * @param {string} targetUrl - The URL to crawl
     * @param {Object} extractors - Custom extraction functions
     * @param {boolean} followSchemaLinks - Whether to follow schema links
     */
    async crawl(targetUrl, extractors = {}, followSchemaLinks = true) {
        try {
            await this.init();
            
            // Fetch the main page
            const pageData = await this.fetchPage(targetUrl);
            
            // Extract schema information from main page
            const mainPageData = this.extractSchemaInfo(pageData.html, extractors);
            mainPageData.url = pageData.url;
            mainPageData.status = pageData.status;

            // Find schema links
            const schemaLinks = this.findSchemaLinks(pageData.html, pageData.url);
            mainPageData.schemaLinks = schemaLinks;

            console.log(`Found ${schemaLinks.length} potential schema links`);

            // Follow schema links if requested
            if (followSchemaLinks && schemaLinks.length > 0) {
                mainPageData.schemaPages = [];
                
                for (const link of schemaLinks.slice(0, 3)) { // Limit to first 3 links
                    try {
                        console.log(`Following schema link: ${link.url}`);
                        const schemaPageData = await this.fetchPage(link.url);
                        const schemaData = this.extractSchemaInfo(schemaPageData.html, extractors);
                        schemaData.url = link.url;
                        schemaData.linkText = link.text;
                        
                        mainPageData.schemaPages.push(schemaData);
                    } catch (error) {
                        console.warn(`Failed to crawl schema link ${link.url}: ${error.message}`);
                    }
                }
            }

            // Save data
            await this.saveData(mainPageData);
            
            return mainPageData;
        } catch (error) {
            console.error(`Crawling failed: ${error.message}`);
            throw error;
        }
    }
}

// Example usage function
async function exampleUsage() {
    const crawler = new SimpleDetailsCrawler({
        outputDir: './simple_crawler_results'
    });

    try {
        // Example: crawl a Vietnamese e-commerce site
        const data = await crawler.crawl('https://tiki.vn', {
            // Custom extractor for Tiki-specific elements
            tikiSpecific: () => {
                // This would contain Tiki-specific extraction logic
                return { message: 'Tiki-specific extraction would go here' };
            }
        });

        console.log('Crawling completed!');
        console.log('Extracted data keys:', Object.keys(data));
    } catch (error) {
        console.error('Error:', error.message);
    }
}

// Export for use as a module
module.exports = SimpleDetailsCrawler;

// If run directly, show usage
if (require.main === module) {
    console.log('Simple Details Crawler - Schema Information Extractor');
    console.log('Usage: node simple_details_crawler.js');
    console.log('This version uses HTTP requests and HTML parsing (no browser required)');
    
    // Uncomment to run example
    // exampleUsage();
}