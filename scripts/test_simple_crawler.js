/**
 * Test the simple details crawler
 */

const SimpleDetailsCrawler = require('./simple_details_crawler');

async function testSimpleCrawler() {
    console.log('Testing Simple Details Crawler...');
    
    const crawler = new SimpleDetailsCrawler({
        outputDir: './test_simple_results',
        timeout: 10000
    });

    try {
        // Test with a simple website that has structured data
        console.log('Testing with example.com...');
        
        const data = await crawler.crawl('http://example.com', {
            // Custom extractor for testing
            customTest: () => {
                return { 
                    testMessage: 'Custom extractor working!',
                    timestamp: new Date().toISOString()
                };
            }
        }, false); // Don't follow schema links for this test

        console.log('✅ Simple crawler test completed successfully!');
        console.log('Extracted data structure:');
        console.log('- Metadata:', data.metadata ? '✅' : '❌');
        console.log('- JSON-LD data:', data.jsonLd ? '✅' : '❌');
        console.log('- Microdata:', data.microdata ? '✅' : '❌');
        console.log('- Tables:', data.tables ? '✅' : '❌');
        console.log('- Custom test:', data.customTest ? '✅' : '❌');
        
        if (data.metadata) {
            console.log('Page title:', data.metadata.title);
            console.log('Page description:', data.metadata.description);
        }

        return true;
    } catch (error) {
        console.error('❌ Simple crawler test failed:', error.message);
        return false;
    }
}

// Test with a mock HTML string (for offline testing)
async function testHtmlParsing() {
    console.log('\nTesting HTML parsing capabilities...');
    
    const SimpleDetailsCrawler = require('./simple_details_crawler');
    const crawler = new SimpleDetailsCrawler();
    
    // Mock HTML with schema data
    const mockHtml = `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Test Product Page</title>
        <meta name="description" content="This is a test product page with schema data">
        <script type="application/ld+json">
        {
            "@context": "https://schema.org",
            "@type": "Product",
            "name": "Test Product",
            "description": "A sample product for testing",
            "offers": {
                "@type": "Offer",
                "price": "99.99",
                "priceCurrency": "USD"
            }
        }
        </script>
    </head>
    <body>
        <div itemscope itemtype="https://schema.org/Product">
            <h1 itemprop="name" class="product-name">Test Product</h1>
            <p itemprop="description" class="product-description">A sample product for testing</p>
            <span itemprop="offers" itemscope itemtype="https://schema.org/Offer">
                <span itemprop="price" class="product-price">$99.99</span>
            </span>
        </div>
        
        <table class="product-specs">
            <thead>
                <tr><th>Property</th><th>Value</th></tr>
            </thead>
            <tbody>
                <tr><td>Brand</td><td>TestBrand</td></tr>
                <tr><td>Model</td><td>TestModel</td></tr>
            </tbody>
        </table>
        
        <div class="schema-info">
            <h2>Schema Information</h2>
            <p>This product follows schema.org Product specification</p>
        </div>
    </body>
    </html>
    `;
    
    try {
        const data = crawler.extractSchemaInfo(mockHtml);
        
        console.log('✅ HTML parsing test completed!');
        console.log('Results:');
        console.log('- JSON-LD found:', data.jsonLd?.length || 0, 'items');
        console.log('- Microdata found:', data.microdata?.length || 0, 'items');
        console.log('- Tables found:', data.tables?.length || 0, 'items');
        console.log('- Schema content found:', data.schemaContent?.length || 0, 'items');
        console.log('- Product info:', data.productInfo?.name || 'Not found');
        
        if (data.jsonLd && data.jsonLd.length > 0) {
            console.log('- JSON-LD Product name:', data.jsonLd[0].name);
        }
        
        if (data.microdata && data.microdata.length > 0) {
            console.log('- Microdata Product name:', data.microdata[0].properties.name);
        }
        
        return true;
    } catch (error) {
        console.error('❌ HTML parsing test failed:', error.message);
        return false;
    }
}

async function runAllTests() {
    console.log('='.repeat(60));
    console.log('SIMPLE DETAILS CRAWLER TEST SUITE');
    console.log('='.repeat(60));
    
    let passedTests = 0;
    let totalTests = 0;
    
    // Test 1: HTML parsing
    totalTests++;
    if (await testHtmlParsing()) {
        passedTests++;
    }
    
    // Test 2: Simple web crawling (if network is available)
    totalTests++;
    if (await testSimpleCrawler()) {
        passedTests++;
    }
    
    console.log('\n' + '='.repeat(60));
    console.log(`TEST RESULTS: ${passedTests}/${totalTests} tests passed`);
    console.log('='.repeat(60));
    
    if (passedTests === totalTests) {
        console.log('🎉 All tests passed! The crawler is working correctly.');
        console.log('\nNext steps:');
        console.log('1. Modify the crawler configuration for your specific website');
        console.log('2. Add custom extractors for your schema requirements');
        console.log('3. Run: node simple_details_crawler.js to start crawling');
    } else {
        console.log('⚠️  Some tests failed. Check the error messages above.');
    }
}

// Run tests if this file is executed directly
if (require.main === module) {
    runAllTests().catch(console.error);
}

module.exports = { testSimpleCrawler, testHtmlParsing, runAllTests };