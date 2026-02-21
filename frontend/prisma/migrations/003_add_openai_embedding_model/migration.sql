INSERT INTO embedding_models (id, name, vector_dimension, provider, parameters, is_active)
VALUES (
    'text-embedding-3-small',
    'text-embedding-3-small',
    1536,
    'openai',
    '{"dimensions":1536}'::jsonb,
    false
)
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    vector_dimension = EXCLUDED.vector_dimension,
    provider = EXCLUDED.provider,
    parameters = EXCLUDED.parameters,
    is_active = EXCLUDED.is_active;
