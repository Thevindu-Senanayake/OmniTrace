/**
 * Gateway configuration, read from the environment. Backend URLs default to the
 * docker-compose service names so no config is needed in the container.
 */
export const config = {
	port: process.env.PORT,
	orderServiceUrl: process.env.ORDER_SERVICE_URL,
	productCatalogUrl: process.env.PRODUCT_CATALOG_URL,
};

export const REQUEST_ID_HEADER = "x-request-id";
