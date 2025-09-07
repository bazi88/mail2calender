/**
 * Configuration file for common crawling scenarios
 * Customize these configurations for your specific needs
 */

// Common Vietnamese websites and their schema tab selectors
const VIETNAMESE_SITES = {
    'tiki.vn': {
        schemaTab: '.product-schema, .schema-tab, a[href*="schema"]',
        extractors: {
            productSchema: async (page) => {
                return await page.evaluate(() => {
                    const product = {
                        name: document.querySelector('h1, .product-name, [data-view-id="pdp_product_name"]')?.textContent?.trim(),
                        price: document.querySelector('.product-price, .price-sale, [data-view-id="pdp_price"]')?.textContent?.trim(),
                        description: document.querySelector('.product-description, .description, [data-view-id="pdp_product_description"]')?.textContent?.trim(),
                        brand: document.querySelector('.brand, .product-brand')?.textContent?.trim(),
                        category: document.querySelector('.breadcrumb, .category')?.textContent?.trim(),
                        images: Array.from(document.querySelectorAll('.product-image img, .gallery img')).map(img => img.src),
                        rating: document.querySelector('.rating, .stars, [data-view-id="pdp_rating"]')?.textContent?.trim(),
                        availability: document.querySelector('.availability, .stock-status')?.textContent?.trim()
                    };
                    return product;
                });
            }
        }
    },
    
    'shopee.vn': {
        schemaTab: '.product-briefing, .schema-info, [data-testid="schema"]',
        extractors: {
            productSchema: async (page) => {
                return await page.evaluate(() => {
                    return {
                        name: document.querySelector('._44qnta, .product-title')?.textContent?.trim(),
                        price: document.querySelector('._16N7-Y, .product-price')?.textContent?.trim(),
                        description: document.querySelector('.irIKAp, .product-detail')?.textContent?.trim(),
                        shop: document.querySelector('._1k06nj, .shop-name')?.textContent?.trim(),
                        rating: document.querySelector('.shopee-rating, .rating')?.textContent?.trim(),
                        sold: document.querySelector('.sold-count')?.textContent?.trim()
                    };
                });
            }
        }
    },
    
    'sendo.vn': {
        schemaTab: '.product-info-schema, .schema-tab',
        extractors: {
            productSchema: async (page) => {
                return await page.evaluate(() => {
                    return {
                        name: document.querySelector('.d-product-name, h1')?.textContent?.trim(),
                        price: document.querySelector('.d-price, .price')?.textContent?.trim(),
                        description: document.querySelector('.d-product-description')?.textContent?.trim(),
                        brand: document.querySelector('.d-brand')?.textContent?.trim(),
                        specifications: Array.from(document.querySelectorAll('.specification-item')).map(item => ({
                            key: item.querySelector('.spec-key')?.textContent?.trim(),
                            value: item.querySelector('.spec-value')?.textContent?.trim()
                        }))
                    };
                });
            }
        }
    }
};

// Common event websites
const EVENT_SITES = {
    'eventbrite.com': {
        schemaTab: '.event-schema, .structured-data-tab',
        extractors: {
            eventSchema: async (page) => {
                return await page.evaluate(() => {
                    return {
                        name: document.querySelector('h1.event-title, [data-automation="event-title"]')?.textContent?.trim(),
                        startDate: document.querySelector('[data-automation="event-start-date"]')?.textContent?.trim(),
                        endDate: document.querySelector('[data-automation="event-end-date"]')?.textContent?.trim(),
                        location: document.querySelector('[data-automation="event-location"]')?.textContent?.trim(),
                        description: document.querySelector('.event-description, [data-automation="event-description"]')?.textContent?.trim(),
                        organizer: document.querySelector('.organizer-name, [data-automation="organizer-name"]')?.textContent?.trim(),
                        price: document.querySelector('.ticket-price, [data-automation="ticket-price"]')?.textContent?.trim()
                    };
                });
            }
        }
    },
    
    'ticketbox.vn': {
        schemaTab: '.event-schema, .lược-đồ',
        extractors: {
            eventSchema: async (page) => {
                return await page.evaluate(() => {
                    return {
                        name: document.querySelector('.event-name, h1')?.textContent?.trim(),
                        date: document.querySelector('.event-date, .date')?.textContent?.trim(),
                        time: document.querySelector('.event-time, .time')?.textContent?.trim(),
                        venue: document.querySelector('.event-venue, .venue')?.textContent?.trim(),
                        city: document.querySelector('.event-city, .city')?.textContent?.trim(),
                        category: document.querySelector('.event-category')?.textContent?.trim(),
                        description: document.querySelector('.event-description')?.textContent?.trim()
                    };
                });
            }
        }
    }
};

// Schema.org specific extractors
const SCHEMA_ORG_EXTRACTORS = {
    // Extract Product schema
    product: async (page) => {
        return await page.evaluate(() => {
            const productData = {};
            
            // Look for JSON-LD structured data
            const scripts = document.querySelectorAll('script[type="application/ld+json"]');
            for (const script of scripts) {
                try {
                    const data = JSON.parse(script.textContent);
                    if (data['@type'] === 'Product' || data['@type']?.includes('Product')) {
                        return data;
                    }
                } catch (e) {
                    // Continue to next script
                }
            }
            
            // Fallback to microdata
            const product = document.querySelector('[itemtype*="Product"]');
            if (product) {
                productData.name = product.querySelector('[itemprop="name"]')?.textContent?.trim();
                productData.description = product.querySelector('[itemprop="description"]')?.textContent?.trim();
                productData.image = product.querySelector('[itemprop="image"]')?.src;
                productData.url = product.querySelector('[itemprop="url"]')?.href;
                
                const offer = product.querySelector('[itemprop="offers"]');
                if (offer) {
                    productData.price = offer.querySelector('[itemprop="price"]')?.textContent?.trim();
                    productData.currency = offer.querySelector('[itemprop="priceCurrency"]')?.textContent?.trim();
                    productData.availability = offer.querySelector('[itemprop="availability"]')?.textContent?.trim();
                }
                
                const brand = product.querySelector('[itemprop="brand"]');
                if (brand) {
                    productData.brand = brand.querySelector('[itemprop="name"]')?.textContent?.trim();
                }
            }
            
            return productData;
        });
    },

    // Extract Event schema
    event: async (page) => {
        return await page.evaluate(() => {
            const eventData = {};
            
            // Look for JSON-LD structured data
            const scripts = document.querySelectorAll('script[type="application/ld+json"]');
            for (const script of scripts) {
                try {
                    const data = JSON.parse(script.textContent);
                    if (data['@type'] === 'Event' || data['@type']?.includes('Event')) {
                        return data;
                    }
                } catch (e) {
                    // Continue to next script
                }
            }
            
            // Fallback to microdata
            const event = document.querySelector('[itemtype*="Event"]');
            if (event) {
                eventData.name = event.querySelector('[itemprop="name"]')?.textContent?.trim();
                eventData.description = event.querySelector('[itemprop="description"]')?.textContent?.trim();
                eventData.startDate = event.querySelector('[itemprop="startDate"]')?.getAttribute('datetime');
                eventData.endDate = event.querySelector('[itemprop="endDate"]')?.getAttribute('datetime');
                eventData.url = event.querySelector('[itemprop="url"]')?.href;
                
                const location = event.querySelector('[itemprop="location"]');
                if (location) {
                    eventData.location = {
                        name: location.querySelector('[itemprop="name"]')?.textContent?.trim(),
                        address: location.querySelector('[itemprop="address"]')?.textContent?.trim()
                    };
                }
                
                const organizer = event.querySelector('[itemprop="organizer"]');
                if (organizer) {
                    eventData.organizer = organizer.querySelector('[itemprop="name"]')?.textContent?.trim();
                }
            }
            
            return eventData;
        });
    },

    // Extract Organization schema
    organization: async (page) => {
        return await page.evaluate(() => {
            const orgData = {};
            
            // Look for JSON-LD structured data
            const scripts = document.querySelectorAll('script[type="application/ld+json"]');
            for (const script of scripts) {
                try {
                    const data = JSON.parse(script.textContent);
                    if (data['@type'] === 'Organization' || data['@type']?.includes('Organization')) {
                        return data;
                    }
                } catch (e) {
                    // Continue to next script
                }
            }
            
            // Fallback to microdata
            const org = document.querySelector('[itemtype*="Organization"]');
            if (org) {
                orgData.name = org.querySelector('[itemprop="name"]')?.textContent?.trim();
                orgData.url = org.querySelector('[itemprop="url"]')?.href;
                orgData.logo = org.querySelector('[itemprop="logo"]')?.src;
                orgData.description = org.querySelector('[itemprop="description"]')?.textContent?.trim();
                orgData.email = org.querySelector('[itemprop="email"]')?.textContent?.trim();
                orgData.telephone = org.querySelector('[itemprop="telephone"]')?.textContent?.trim();
                
                const address = org.querySelector('[itemprop="address"]');
                if (address) {
                    orgData.address = {
                        streetAddress: address.querySelector('[itemprop="streetAddress"]')?.textContent?.trim(),
                        addressLocality: address.querySelector('[itemprop="addressLocality"]')?.textContent?.trim(),
                        addressRegion: address.querySelector('[itemprop="addressRegion"]')?.textContent?.trim(),
                        postalCode: address.querySelector('[itemprop="postalCode"]')?.textContent?.trim(),
                        addressCountry: address.querySelector('[itemprop="addressCountry"]')?.textContent?.trim()
                    };
                }
            }
            
            return orgData;
        });
    }
};

// Common tab selectors for different languages
const TAB_SELECTORS = {
    vietnamese: [
        'a:contains("Lược đồ")',
        'button:contains("Lược đồ")',
        '.tab:contains("Lược đồ")',
        '[data-tab="luoc-do"]',
        '[data-tab="schema"]',
        '.schema-tab-vn'
    ],
    english: [
        'a:contains("Schema")',
        'button:contains("Schema")',
        '.tab:contains("Schema")',
        '[data-tab="schema"]',
        '.schema-tab',
        'a[href*="schema"]'
    ],
    chinese: [
        'a:contains("架构")',
        'button:contains("架构")',
        '.tab:contains("架构")',
        'a:contains("模式")',
        'button:contains("模式")'
    ]
};

module.exports = {
    VIETNAMESE_SITES,
    EVENT_SITES,
    SCHEMA_ORG_EXTRACTORS,
    TAB_SELECTORS
};