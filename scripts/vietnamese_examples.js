/**
 * Example usage of Details Crawler for Vietnamese websites
 * This file demonstrates how to use the crawler for common Vietnamese scenarios
 */

const SimpleDetailsCrawler = require('./simple_details_crawler');
const { VIETNAMESE_SITES, SCHEMA_ORG_EXTRACTORS } = require('./crawler_config');

/**
 * Example 1: Crawl Vietnamese e-commerce product pages
 */
async function crawlVietnameseEcommerce() {
    console.log('🛒 Crawling Vietnamese E-commerce Example');
    console.log('=' .repeat(50));
    
    const crawler = new SimpleDetailsCrawler({
        outputDir: './ecommerce_results'
    });

    // Example product URLs (replace with actual URLs)
    const productUrls = [
        // 'https://tiki.vn/product-example',
        // 'https://shopee.vn/product-example',
        // 'https://sendo.vn/product-example'
    ];

    // Custom extractor for Vietnamese e-commerce
    const ecommerceExtractor = {
        vietnameseProduct: () => {
            // This function would extract Vietnamese product information
            return {
                name: 'Sản phẩm mẫu', // Sample product name
                price: '100.000 VND',
                description: 'Mô tả sản phẩm bằng tiếng Việt',
                category: 'Danh mục sản phẩm',
                brand: 'Thương hiệu',
                specifications: {
                    'Kích thước': '10x10x10 cm',
                    'Trọng lượng': '500g',
                    'Màu sắc': 'Đỏ, Xanh, Vàng'
                }
            };
        },
        
        priceHistory: () => {
            // Extract price and discount information
            return {
                originalPrice: '120.000 VND',
                salePrice: '100.000 VND',
                discount: '20.000 VND (17%)',
                promotions: ['Miễn phí vận chuyển', 'Giảm 10% khi mua 2']
            };
        },
        
        sellerInfo: () => {
            // Extract seller/shop information
            return {
                shopName: 'Cửa hàng ABC',
                rating: '4.8/5',
                followerCount: '10.5K',
                location: 'TP.HCM',
                responseRate: '95%',
                joinDate: '2020-01-15'
            };
        }
    };

    console.log('This example shows how to extract Vietnamese e-commerce data');
    console.log('Replace the productUrls array with actual Vietnamese e-commerce URLs');
    console.log('The extractor functions above show the type of data that can be extracted');
    
    return ecommerceExtractor;
}

/**
 * Example 2: Crawl Vietnamese event websites
 */
async function crawlVietnameseEvents() {
    console.log('\n🎪 Crawling Vietnamese Events Example');
    console.log('=' .repeat(50));
    
    const crawler = new SimpleDetailsCrawler({
        outputDir: './events_results'
    });

    // Custom extractor for Vietnamese events
    const eventExtractor = {
        vietnameseEvent: () => {
            return {
                name: 'Sự kiện Âm nhạc Việt Nam 2024',
                date: '15/12/2024',
                time: '19:00 - 23:00',
                venue: 'Nhà hát Thành phố',
                address: '123 Đường Nguyễn Huệ, Quận 1, TP.HCM',
                ticketTypes: [
                    { type: 'VIP', price: '500.000 VND', available: true },
                    { type: 'Thường', price: '200.000 VND', available: true },
                    { type: 'Sinh viên', price: '100.000 VND', available: false }
                ],
                organizer: 'Công ty Tổ chức Sự kiện ABC',
                category: 'Âm nhạc',
                ageLimit: '16+',
                description: 'Đêm nhạc quy tụ các ca sĩ nổi tiếng Việt Nam'
            };
        },
        
        eventSchedule: () => {
            return {
                doors: '18:00',
                soundcheck: '18:30',
                showStart: '19:00',
                intermission: '20:30',
                showEnd: '23:00',
                schedule: [
                    { time: '19:00', activity: 'Khai mạc' },
                    { time: '19:15', activity: 'Ca sĩ A biểu diễn' },
                    { time: '20:00', activity: 'Ca sĩ B biểu diễn' },
                    { time: '20:30', activity: 'Giải lao' },
                    { time: '21:00', activity: 'Ca sĩ C biểu diễn' },
                    { time: '22:30', activity: 'Bế mạc' }
                ]
            };
        }
    };

    console.log('This example shows how to extract Vietnamese event data');
    console.log('Common Vietnamese event information includes:');
    console.log('- Event name, date, time, venue');
    console.log('- Ticket types and prices in VND');
    console.log('- Vietnamese address format');
    console.log('- Age restrictions and categories');
    
    return eventExtractor;
}

/**
 * Example 3: Extract schema data from Vietnamese business websites
 */
async function crawlVietnameseBusiness() {
    console.log('\n🏢 Crawling Vietnamese Business Example');
    console.log('=' .repeat(50));
    
    const businessExtractor = {
        vietnameseBusiness: () => {
            return {
                companyName: 'Công ty TNHH ABC',
                businessType: 'Công ty Trách nhiệm Hữu hạn',
                taxCode: '0123456789',
                address: {
                    street: '123 Đường Lê Lợi',
                    ward: 'Phường Bến Thành',
                    district: 'Quận 1',
                    city: 'TP. Hồ Chí Minh',
                    postalCode: '70000'
                },
                contact: {
                    phone: '028-1234-5678',
                    email: 'info@abc.com.vn',
                    website: 'https://abc.com.vn',
                    fax: '028-1234-5679'
                },
                businessInfo: {
                    establishedDate: '15/01/2010',
                    capital: '5.000.000.000 VND',
                    employees: '50-100 người',
                    industry: 'Công nghệ thông tin',
                    mainProducts: ['Phần mềm quản lý', 'Ứng dụng di động']
                }
            };
        },
        
        certifications: () => {
            return {
                businessLicense: 'Giấy phép kinh doanh số 0123456789',
                qualityCerts: ['ISO 9001:2015', 'ISO 27001:2013'],
                awards: ['Top 10 công ty IT Việt Nam 2023'],
                memberships: ['Hiệp hội Phần mềm Việt Nam', 'VINASA']
            };
        }
    };

    console.log('This example shows how to extract Vietnamese business data');
    console.log('Includes Vietnamese-specific business information like:');
    console.log('- Vietnamese company types (TNHH, Cổ phần, etc.)');
    console.log('- Tax codes and business licenses');
    console.log('- Vietnamese address format');
    console.log('- Capital in VND currency');
    
    return businessExtractor;
}

/**
 * Example 4: Generic schema extraction for any Vietnamese website
 */
async function extractVietnameseSchema(url) {
    console.log(`\n📊 Extracting Schema from: ${url}`);
    console.log('=' .repeat(50));
    
    const crawler = new SimpleDetailsCrawler({
        outputDir: './vietnamese_schema_results'
    });

    try {
        const data = await crawler.crawl(url, {
            // Vietnamese-specific content extractor
            vietnameseContent: () => {
                // This would analyze the page for Vietnamese schema content
                return {
                    language: 'vi-VN',
                    schemaKeywords: [
                        'lược đồ', 'cấu trúc dữ liệu', 'thông tin sản phẩm',
                        'chi tiết sự kiện', 'thông tin doanh nghiệp'
                    ],
                    vietnameseFields: {
                        'Tên': 'Name field in Vietnamese',
                        'Mô tả': 'Description field in Vietnamese', 
                        'Giá': 'Price field in Vietnamese',
                        'Địa chỉ': 'Address field in Vietnamese',
                        'Thời gian': 'Time field in Vietnamese'
                    }
                };
            },

            // Use pre-configured schema extractors
            ...SCHEMA_ORG_EXTRACTORS
        });

        console.log('✅ Schema extraction completed');
        console.log('Found the following schema types:');
        
        if (data.jsonLd && data.jsonLd.length > 0) {
            data.jsonLd.forEach((schema, index) => {
                console.log(`  ${index + 1}. ${schema['@type']} - ${schema.name || 'Unnamed'}`);
            });
        }
        
        return data;
    } catch (error) {
        console.error('❌ Schema extraction failed:', error.message);
        return null;
    }
}

/**
 * Main demonstration function
 */
async function runVietnameseExamples() {
    console.log('🇻🇳 VIETNAMESE WEBSITE CRAWLER EXAMPLES');
    console.log('=' .repeat(60));
    console.log('This file demonstrates how to use the crawler for Vietnamese websites');
    console.log('Each example shows different extraction strategies for common use cases\n');

    try {
        // Run all examples
        await crawlVietnameseEcommerce();
        await crawlVietnameseEvents();
        await crawlVietnameseBusiness();
        
        console.log('\n📝 Usage Instructions:');
        console.log('1. Replace example URLs with real Vietnamese website URLs');
        console.log('2. Modify the extractor functions to match your target website structure');
        console.log('3. Run: node vietnamese_examples.js to start crawling');
        console.log('4. Check the output directories for extracted data');
        
        console.log('\n🔧 Customization Tips:');
        console.log('- Use browser developer tools to find the right CSS selectors');
        console.log('- Test with headless: false to see what the crawler is doing');
        console.log('- Add delays if the website loads content dynamically');
        console.log('- Use the crawler_config.js file for site-specific configurations');
        
    } catch (error) {
        console.error('Example execution failed:', error);
    }
}

// Export functions for use in other files
module.exports = {
    crawlVietnameseEcommerce,
    crawlVietnameseEvents,
    crawlVietnameseBusiness,
    extractVietnameseSchema,
    runVietnameseExamples
};

// Run examples if this file is executed directly
if (require.main === module) {
    runVietnameseExamples();
}