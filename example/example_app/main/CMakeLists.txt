idf_component_register(
    SRCS "main.c"
    INCLUDE_DIRS ""
    REQUIRES ulp
)

set(ulp_app_name ulp_${COMPONENT_NAME})
set(ulp_s_sources "out.S")
set(ulp_exp_dep_srcs "main.c")
ulp_embed_binary(${ulp_app_name} "${ulp_s_sources}" "${ulp_exp_dep_srcs}")
