/**
 * Test file for Details Crawler
 * This demonstrates how to use the crawler to extract schema information
 */

const DetailsCrawler = require('./details_crawler');

async function testCrawler() {
    console.log('Starting Details Crawler Test...');
    
    const crawler = new DetailsCrawler({
        headless: true,  // Set to false if you want to see the browser
        outputDir: './test_results',
        timeout: 30000
    });

    try {
        // Test with a sample website that might have schema information
        console.log('Testing with schema.org example...');
        
        const data = await crawler.crawl('https://schema.org/Event', null, {
            // Custom extractor for Event schema properties
            eventProperties: async () => {
                return await crawler.page.evaluate(() => {
                    const properties = document.querySelectorAll('.prop-name, .property-name, dt');
                    return Array.from(properties).map(prop => {
                        const name = prop.textContent.trim();
                        const description = prop.nextElementSibling?.textContent.trim() || '';
                        return { name, description };
                    }).filter(prop => prop.name);
                });
            },
            
            // Extract schema markup examples
            schemaExamples: async () => {
                return await crawler.page.evaluate(() => {
                    const examples = document.querySelectorAll('pre, code, .example');
                    return Array.from(examples).map(example => ({
                        type: example.tagName.toLowerCase(),
                        content: example.textContent.trim()
                    })).filter(ex => ex.content.length > 0);
                });
            }
        });

        console.log('Crawling completed successfully!');
        console.log('Sample of extracted data:');
        console.log('- Metadata:', data.metadata);
        console.log('- Number of tables found:', data.tables?.length || 0);
        console.log('- Schema text elements found:', data.schemaText?.length || 0);
        console.log('- Event properties found:', data.eventProperties?.length || 0);
        
    } catch (error) {
        console.error('Test failed:', error.message);
    }
}

// Test with Vietnamese website example
async function testVietnameseWebsite() {
    console.log('Testing with Vietnamese website...');
    
    const crawler = new DetailsCrawler({
        headless: true,
        outputDir: './test_results_vn'
    });

    try {
        // Example: Test with a Vietnamese e-commerce or event website
        // This is just an example - replace with actual URL that has schema tab
        const data = await crawler.crawl('https://example.vn', 
            // Try to find schema tab in Vietnamese
            'a:contains("Lược đồ"), button:contains("Lược đồ"), .tab:contains("Schema")',
            {
                // Custom extractor for Vietnamese content
                vietnameseContent: async () => {
                    return await crawler.page.evaluate(() => {
                        // Look for Vietnamese schema-related keywords
                        const keywords = ['lược đồ', 'cấu trúc', 'mô tả', 'thuộc tính', 'dữ liệu'];
                        const content = document.body.textContent.toLowerCase();
                        const foundKeywords = keywords.filter(keyword => content.includes(keyword));
                        
                        return {
                            foundKeywords,
                            hasVietnameseSchema: foundKeywords.length > 0
                        };
                    });
                },
                
                // Extract product or event information if available
                productInfo: async () => {
                    return await crawler.page.evaluate(() => {
                        const products = document.querySelectorAll(
                            '.product, .event, .item, [itemtype*="Product"], [itemtype*="Event"]'
                        );
                        return Array.from(products).map(product => ({
                            name: product.querySelector('h1, h2, h3, .name, .title')?.textContent.trim(),
                            price: product.querySelector('.price, .cost, [itemprop="price"]')?.textContent.trim(),
                            description: product.querySelector('.description, .desc, [itemprop="description"]')?.textContent.trim()
                        })).filter(p => p.name);
                    });
                }
            }
        );

        console.log('Vietnamese website test completed!');
        console.log('Data extracted:', Object.keys(data));
        
    } catch (error) {
        console.error('Vietnamese website test failed:', error.message);
        console.log('This is expected if the example URL is not accessible');
    }
}

// Main test function
async function runTests() {
    console.log('='.repeat(50));
    console.log('DETAILS CRAWLER TEST SUITE');
    console.log('='.repeat(50));
    
    try {
        await testCrawler();
        console.log('\n' + '-'.repeat(30) + '\n');
        await testVietnameseWebsite();
    } catch (error) {
        console.error('Test suite failed:', error);
    }
    
    console.log('\n' + '='.repeat(50));
    console.log('TEST SUITE COMPLETED');
    console.log('='.repeat(50));
}

// Run tests if this file is executed directly
if (require.main === module) {
    runTests();
}

module.exports = { testCrawler, testVietnameseWebsite, runTests };