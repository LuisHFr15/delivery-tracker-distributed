#!/usr/bin/env bash
# ==============================================================================
# Deploy Docker Images to AWS ECR and Update AWS Lambda Functions
# Delivery Tracker Distributed Pipeline
# ==============================================================================

set -euo pipefail

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Base directory (project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${PROJECT_ROOT}"

# Default configurations
SERVICES_AVAILABLE=("ingester" "processor" "notifier")
SERVICE=""
TAG=""
UPDATE_LAMBDA=false
CUSTOM_LAMBDA_NAME=""
AWS_REGION="${AWS_REGION:-${AWS_DEFAULT_REGION:-}}"
AWS_ACCOUNT_ID="${AWS_ACCOUNT_ID:-}"
AWS_PROFILE="${AWS_PROFILE:-}"
ECR_PREFIX="delivery-tracker"
PLATFORM="linux/amd64"
AUTO_CREATE_REPO=true

# Helper: Print styled messages
log_info()    { echo -e "${BLUE}${BOLD}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}${BOLD}[SUCCESS]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}${BOLD}[WARN]${NC} $1"; }
log_error()   { echo -e "${RED}${BOLD}[ERROR]${NC} $1" >&2; }
log_header()  {
    echo -e "\n${PURPLE}${BOLD}=================================================================${NC}"
    echo -e "${PURPLE}${BOLD}  $1${NC}"
    echo -e "${PURPLE}${BOLD}=================================================================${NC}\n"
}

# Print help / usage
show_help() {
    echo -e "${BOLD}Delivery Tracker — ECR Deploy & Lambda Update Script${NC}"
    echo ""
    echo -e "${BOLD}USO:${NC}"
    echo "  ./scripts/deploy-ecr.sh [SERVIÇO] [TAG] [OPÇÕES]"
    echo ""
    echo -e "${BOLD}SERVIÇOS DISPONÍVEIS:${NC}"
    echo "  ingester, processor, notifier, all (padrão: all)"
    echo ""
    echo -e "${BOLD}EXEMPLOS RÁPIDOS (com poucos parâmetros):${NC}"
    echo -e "  ${CYAN}# 1. Deploy de todos os serviços no ECR com tag automática:${NC}"
    echo "  ./scripts/deploy-ecr.sh"
    echo ""
    echo -e "  ${CYAN}# 2. Deploy de apenas um serviço no ECR:${NC}"
    echo "  ./scripts/deploy-ecr.sh ingester"
    echo ""
    echo -e "  ${CYAN}# 3. Deploy de um serviço e atualização imediata da Lambda correspondente:${NC}"
    echo "  ./scripts/deploy-ecr.sh ingester -l"
    echo ""
    echo -e "  ${CYAN}# 4. Deploy de todos os serviços no ECR e atualização de todas as Lambdas:${NC}"
    echo "  ./scripts/deploy-ecr.sh all -l"
    echo ""
    echo -e "  ${CYAN}# 5. Deploy especificando uma tag customizada:${NC}"
    echo "  ./scripts/deploy-ecr.sh processor v1.2.0 -l"
    echo ""
    echo -e "${BOLD}OPÇÕES DETALHADAS:${NC}"
    echo "  -s, --service <nome>     Nome do serviço (ingester, processor, notifier, all)"
    echo "  -t, --tag <tag>          Tag da imagem Docker (padrão: git commit hash curto ou timestamp)"
    echo "  -l, --update-lambda      Atualiza o código da função Lambda com a nova imagem após o push"
    echo "  --lambda-name <nome>     Sobrescreve o nome da função Lambda (apenas para serviço único)"
    echo "  -r, --region <região>    Região AWS (padrão: detectada via AWS CLI ou us-east-1)"
    echo "  -a, --account-id <id>    AWS Account ID (padrão: detectado automaticamente via STS)"
    echo "  -p, --profile <perfil>   AWS CLI Profile a ser utilizado (padrão: \$AWS_PROFILE ou default)"
    echo "  --prefix <prefixo>       Prefixo para o repositório ECR (padrão: delivery-tracker)"
    echo "                           (Ex: delivery-tracker-ingester. Use \"\" para desativar prefixo)"
    echo "  --platform <plataforma>  Plataforma para o build Docker (padrão: linux/amd64)"
    echo "  --no-create-repo         Não cria automaticamente o repositório ECR caso não exista"
    echo "  -h, --help               Exibe esta mensagem de ajuda"
    echo ""
}

# Parse command line arguments
POSITIONAL_ARGS=()
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        -s|--service)
            SERVICE="$2"
            shift 2
            ;;
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -l|--lambda|--update-lambda)
            UPDATE_LAMBDA=true
            shift
            ;;
        --lambda-name)
            CUSTOM_LAMBDA_NAME="$2"
            shift 2
            ;;
        -r|--region)
            AWS_REGION="$2"
            shift 2
            ;;
        -a|--account-id)
            AWS_ACCOUNT_ID="$2"
            shift 2
            ;;
        -p|--profile)
            AWS_PROFILE="$2"
            shift 2
            ;;
        --prefix)
            ECR_PREFIX="$2"
            shift 2
            ;;
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --no-create-repo)
            AUTO_CREATE_REPO=false
            shift
            ;;
        -*)
            log_error "Opção desconhecida: $1"
            show_help
            exit 1
            ;;
        *)
            POSITIONAL_ARGS+=("$1")
            shift
            ;;
    esac
done

# Handle positional arguments
if [[ ${#POSITIONAL_ARGS[@]} -gt 0 ]]; then
    if [[ -z "$SERVICE" ]]; then
        SERVICE="${POSITIONAL_ARGS[0]}"
    fi
fi

if [[ ${#POSITIONAL_ARGS[@]} -gt 1 ]]; then
    if [[ -z "$TAG" ]]; then
        TAG="${POSITIONAL_ARGS[1]}"
    fi
fi

# Set default service if still empty
if [[ -z "$SERVICE" ]]; then
    SERVICE="all"
fi

# Set default tag if still empty (use git commit sha or date)
if [[ -z "$TAG" ]]; then
    if git rev-parse --short HEAD >/dev/null 2>&1; then
        TAG="$(git rev-parse --short HEAD)"
    else
        TAG="$(date +%Y%m%d%H%M%S)"
    fi
fi

# Validate dependencies
check_dependencies() {
    log_info "Verificando ferramentas necessárias..."
    if ! command -v docker &> /dev/null; then
        log_error "Docker não está instalado ou não está no PATH."
        exit 1
    fi

    if ! docker info >/dev/null 2>&1; then
        log_error "O daemon do Docker não está rodando. Por favor, inicie o Docker."
        exit 1
    fi

    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI não está instalado ou não está no PATH."
        exit 1
    fi
}

# Setup AWS profile argument array
AWS_CLI_ARGS=()
if [[ -n "$AWS_PROFILE" ]]; then
    AWS_CLI_ARGS+=(--profile "$AWS_PROFILE")
fi

# Auto-detect AWS Region and Account ID
resolve_aws_context() {
    log_info "Identificando credenciais e região da AWS..."

    # Detect Region
    if [[ -z "$AWS_REGION" ]]; then
        AWS_REGION=$(aws configure get region "${AWS_CLI_ARGS[@]}" 2>/dev/null || echo "")
        if [[ -z "$AWS_REGION" ]]; then
            AWS_REGION="us-east-1"
            log_warn "Região AWS não especificada nem configurada no CLI. Usando padrão: ${AWS_REGION}"
        fi
    fi

    # Detect Account ID
    if [[ -z "$AWS_ACCOUNT_ID" ]]; then
        log_info "Obtendo AWS Account ID via STS..."
        AWS_ACCOUNT_ID=$(aws sts get-caller-identity "${AWS_CLI_ARGS[@]}" --query "Account" --output text 2>/dev/null || echo "")
        if [[ -z "$AWS_ACCOUNT_ID" ]]; then
            log_error "Não foi possível obter o AWS Account ID. Verifique suas credenciais da AWS (aws configure / AWS_PROFILE)."
            exit 1
        fi
    fi

    log_success "AWS Context -> Conta: ${BOLD}${AWS_ACCOUNT_ID}${NC} | Região: ${BOLD}${AWS_REGION}${NC} ${AWS_PROFILE:+| Perfil: ${BOLD}${AWS_PROFILE}${NC}}"
}

# Authenticate Docker with AWS ECR
ecr_login() {
    local ecr_registry="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
    log_info "Autenticando Docker no Amazon ECR (${ecr_registry})..."
    
    if aws ecr get-login-password --region "${AWS_REGION}" "${AWS_CLI_ARGS[@]}" | \
       docker login --username AWS --password-stdin "${ecr_registry}" >/dev/null 2>&1; then
        log_success "Docker autenticado com sucesso no ECR!"
    else
        log_error "Falha na autenticação com o ECR. Verifique suas permissões IAM de ECR."
        exit 1
    fi
}

# Ensure ECR repository exists
ensure_ecr_repo() {
    local repo_name="$1"
    if ! aws ecr describe-repositories --repository-names "${repo_name}" --region "${AWS_REGION}" "${AWS_CLI_ARGS[@]}" >/dev/null 2>&1; then
        if [[ "$AUTO_CREATE_REPO" == true ]]; then
            log_warn "Repositório ECR '${repo_name}' não existe. Criando..."
            aws ecr create-repository \
                --repository-name "${repo_name}" \
                --region "${AWS_REGION}" \
                --image-scanning-configuration scanOnPush=true \
                "${AWS_CLI_ARGS[@]}" >/dev/null
            log_success "Repositório ECR '${repo_name}' criado com sucesso."
        else
            log_error "Repositório ECR '${repo_name}' não existe e --no-create-repo foi informado."
            exit 1
        fi
    fi
}

# Resolve Lambda Function Name for a given service
get_lambda_function_name() {
    local svc="$1"
    if [[ -n "$CUSTOM_LAMBDA_NAME" && "$SERVICE" != "all" ]]; then
        echo "$CUSTOM_LAMBDA_NAME"
        return
    fi

    # Check for specific environment variables (e.g. LAMBDA_NAME_INGESTER)
    local env_var_name="LAMBDA_NAME_${svc^^}"
    if [[ -n "${!env_var_name:-}" ]]; then
        echo "${!env_var_name}"
        return
    fi

    # Default convention based on SAM template / standard naming
    case "$svc" in
        ingester)  echo "IngesterFunction" ;;
        processor) echo "ProcessorFunction" ;;
        notifier)  echo "NotifierFunction" ;;
        *)         echo "${svc}-function" ;;
    esac
}

# Update Lambda function code
update_lambda_code() {
    local svc="$1"
    local image_uri="$2"
    local lambda_name
    lambda_name=$(get_lambda_function_name "$svc")

    log_info "Atualizando código da função Lambda '${BOLD}${lambda_name}${NC}' com a imagem:"
    echo "       ${CYAN}${image_uri}${NC}"

    # Check if function exists
    if ! aws lambda get-function --function-name "${lambda_name}" --region "${AWS_REGION}" "${AWS_CLI_ARGS[@]}" >/dev/null 2>&1; then
        # Try alternate naming: delivery-tracker-<svc> or prefix-<svc>
        local alt_name="${ECR_PREFIX:+${ECR_PREFIX}-}${svc}"
        if aws lambda get-function --function-name "${alt_name}" --region "${AWS_REGION}" "${AWS_CLI_ARGS[@]}" >/dev/null 2>&1; then
            lambda_name="$alt_name"
        else
            log_warn "Função Lambda '${lambda_name}' não foi encontrada na região ${AWS_REGION}."
            log_warn "Se o nome da função for diferente, passe com --lambda-name ou defina LAMBDA_NAME_${svc^^}."
            return 1
        fi
    fi

    # Update function code
    local update_res
    update_res=$(aws lambda update-function-code \
        --function-name "${lambda_name}" \
        --image-uri "${image_uri}" \
        --region "${AWS_REGION}" \
        "${AWS_CLI_ARGS[@]}" 2>&1)

    if [[ $? -ne 0 ]]; then
        log_error "Erro ao atualizar Lambda ${lambda_name}: ${update_res}"
        return 1
    fi

    log_info "Aguardando atualização da Lambda '${lambda_name}' ser concluída..."
    aws lambda wait function-updated \
        --function-name "${lambda_name}" \
        --region "${AWS_REGION}" \
        "${AWS_CLI_ARGS[@]}" 2>/dev/null || true

    log_success "Função Lambda '${BOLD}${lambda_name}${NC}' atualizada com sucesso!"
    return 0
}

# Build, push and deploy a single service
deploy_service() {
    local svc="$1"
    local dockerfile="cmd/${svc}/Dockerfile"

    if [[ ! -f "$dockerfile" ]]; then
        log_error "Dockerfile não encontrado em: ${dockerfile}"
        return 1
    fi

    log_header "Processando Serviço: ${svc}"

    # Construct repository name (e.g. delivery-tracker-notifier)
    local repo_name
    if [[ -n "$ECR_PREFIX" ]]; then
        repo_name="${ECR_PREFIX}-${svc}"
    else
        repo_name="${svc}"
    fi

    local registry="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
    local image_tag_uri="${registry}/${repo_name}:${TAG}"
    local image_latest_uri="${registry}/${repo_name}:latest"

    # Step 1: Ensure ECR Repository exists
    ensure_ecr_repo "${repo_name}"

    # Step 2: Build Docker Image
    log_info "Construindo imagem Docker para '${svc}' (Plataforma: ${PLATFORM})..."
    echo "       Tag: ${CYAN}${TAG}${NC}"
    docker build \
        --platform "${PLATFORM}" \
        --provenance=false \
        -f "${dockerfile}" \
        -t "${repo_name}" \
        .

    log_success "Build concluído com sucesso para ${svc}!"

    # Step 3: Push Docker Image to ECR
    log_info "Enviando imagens para o ECR..."
    docker tag "${repo_name}:latest" \
      "${image_latest_uri}"

    docker push "${image_latest_uri}"
    log_success "Push para o ECR concluído: ${image_tag_uri}"

    # Step 4: Update Lambda (Optional)
    if [[ "$UPDATE_LAMBDA" == true ]]; then
        update_lambda_code "$svc" "$image_latest_uri" || log_warn "Não foi possível atualizar a Lambda para ${svc}."
    fi

    SUMMARY_LIST+=("${svc}|${image_latest_uri}|$( [[ "$UPDATE_LAMBDA" == true ]] && echo "Sim" || echo "Não" )")
}

# Main Execution Flow
main() {
    log_header "Delivery Tracker — Pipeline de Deploy ECR & Lambda"

    check_dependencies
    resolve_aws_context
    ecr_login

    SUMMARY_LIST=()

    local target_services=()
    if [[ "$SERVICE" == "all" ]]; then
        target_services=("${SERVICES_AVAILABLE[@]}")
    else
        # Validate selected service
        local found=false
        for s in "${SERVICES_AVAILABLE[@]}"; do
            if [[ "$s" == "$SERVICE" ]]; then
                found=true
                break
            fi
        done

        if [[ "$found" == false ]]; then
            log_error "Serviço inválido: '${SERVICE}'. Opções válidas: ${SERVICES_AVAILABLE[*]} ou 'all'"
            exit 1
        fi
        target_services=("$SERVICE")
    fi

    # Deploy each targeted service
    for svc in "${target_services[@]}"; do
        deploy_service "$svc"
    done

    # Print final summary
    log_header "Resumo da Execução"
    printf "%-12s | %-15s | %s\n" "SERVIÇO" "LAMBDA ATUALIZADA" "ECR IMAGE URI"
    echo "----------------------------------------------------------------------------------------"
    for item in "${SUMMARY_LIST[@]}"; do
        IFS="|" read -r s_name s_uri s_lambda <<< "$item"
        printf "%-12s | %-17s | %s\n" "${s_name}" "${s_lambda}" "${s_uri}"
    done
    echo "----------------------------------------------------------------------------------------"
    log_success "Processo finalizado com sucesso!\n"
}

main "$@"
